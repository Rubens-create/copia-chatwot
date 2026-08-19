package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/repository"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/whatsapp"
	"github.com/google/uuid"
)

type WhatsAppService interface {
	ProcessWebhookPayload(ctx context.Context, rawBody []byte) error
}

type whatsAppService struct {
	cfg              *config.Config
	db               *database.DB
	accountRepo      repository.AccountRepository
	contactRepo      repository.ContactRepository
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
	eventRepo        repository.EventRepository
	queue            *queue.RedisQueue
}

func NewWhatsAppService(
	cfg *config.Config,
	db *database.DB,
	accountRepo repository.AccountRepository,
	contactRepo repository.ContactRepository,
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
	eventRepo repository.EventRepository,
	queue *queue.RedisQueue,
) WhatsAppService {
	return &whatsAppService{
		cfg:              cfg,
		db:               db,
		accountRepo:      accountRepo,
		contactRepo:      contactRepo,
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		eventRepo:        eventRepo,
		queue:            queue,
	}
}

func (s *whatsAppService) ProcessWebhookPayload(ctx context.Context, rawBody []byte) error {
	parsed, err := whatsapp.ParseWebhookPayload(rawBody)
	if err != nil {
		return fmt.Errorf("failed to parse meta payload: %w", err)
	}

	// 1. Process incoming messages
	for _, msg := range parsed.Messages {
		if err := s.processIncomingMessage(ctx, msg, rawBody); err != nil {
			log.Printf("[WhatsAppService] Error processing message %s: %v", msg.MetaMessageID, err)
		}
	}

	// 2. Process status updates
	for _, status := range parsed.Statuses {
		if err := s.processStatusUpdate(ctx, status, rawBody); err != nil {
			log.Printf("[WhatsAppService] Error processing status %s: %v", status.MetaMessageID, err)
		}
	}

	return nil
}

func (s *whatsAppService) processIncomingMessage(ctx context.Context, msg whatsapp.ParsedIncomingMessage, rawBody []byte) error {
	// 1. Resolve Account and Channel / Inbox before transaction
	var inbox *model.Inbox
	accountID := s.cfg.DefaultAccountID
	inboxID := s.cfg.DefaultInboxID

	if channel, err := s.accountRepo.FindChannelByPhoneOrConfig(ctx, msg.DisplayPhone, msg.PhoneNumberID); err == nil && channel != nil {
		accountID = channel.AccountID
		if inb, err := s.accountRepo.FindInboxByChannelID(ctx, int(channel.ID)); err == nil && inb != nil {
			inboxID = inb.ID
			inbox = inb
		}
	}
	if inbox == nil {
		inbox, _ = s.accountRepo.FindInboxByID(ctx, inboxID)
	}

	// 2. Start Atomic SQL Transaction for the entire incoming message flow
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin atomic transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Idempotency Check inside Transaction
	if msg.MetaMessageID != "" {
		existing, err := s.messageRepo.FindBySourceIDTx(ctx, tx, msg.MetaMessageID)
		if err == nil && existing != nil {
			log.Printf("[WhatsAppService] Message %s already processed. Skipping (Idempotency).", msg.MetaMessageID)
			return nil
		}
	}

	// 4. Find or Create Contact in Tx
	contact, err := s.contactRepo.FindOrCreateByPhoneTx(ctx, tx, accountID, msg.SenderPhone, msg.SenderName)
	if err != nil {
		return fmt.Errorf("failed to find/create contact in tx: %w", err)
	}

	// 5. Find or Create ContactInbox in Tx
	contactInbox, err := s.contactRepo.FindOrCreateContactInboxTx(ctx, tx, int64(contact.ID), int64(inboxID), msg.SenderPhone)
	if err != nil {
		return fmt.Errorf("failed to find/create contact_inbox in tx: %w", err)
	}

	// 6. Find or Create/Reopen Conversation in Tx (Respecting lock_to_single_conversation)
	var conv *model.Conversation
	isNewConversation := false

	if inbox != nil && inbox.LockToSingleConversation {
		// If lock_to_single_conversation is true, find the latest conversation for this contact
		conv, err = s.conversationRepo.FindLatestByContactAndInboxTx(ctx, tx, accountID, inboxID, int64(contact.ID))
		if err != nil {
			return fmt.Errorf("failed to find latest conversation: %w", err)
		}
		if conv != nil {
			if conv.Status != model.ConversationStatusOpen {
				// Reopen resolved conversation
				if err := s.conversationRepo.ReopenTx(ctx, tx, conv.ID); err != nil {
					return fmt.Errorf("failed to reopen conversation in tx: %w", err)
				}
				conv.Status = model.ConversationStatusOpen
			} else {
				_ = s.conversationRepo.UpdateLastActivityTx(ctx, tx, conv.ID)
			}
		}
	} else {
		// Standard flow: search for an open conversation
		conv, err = s.conversationRepo.FindOpenByContactAndInboxTx(ctx, tx, accountID, inboxID, int64(contact.ID))
		if err != nil {
			return fmt.Errorf("failed to find open conversation: %w", err)
		}
		if conv != nil {
			_ = s.conversationRepo.UpdateLastActivityTx(ctx, tx, conv.ID)
		}
	}

	// If no existing or reopenable conversation found, create a new one
	if conv == nil {
		isNewConversation = true
		newConv := &model.Conversation{
			AccountID:            accountID,
			InboxID:              inboxID,
			ContactID:            int64(contact.ID),
			ContactInboxID:       contactInbox.ID,
			Status:               model.ConversationStatusOpen,
			UUID:                 uuid.New().String(),
			LastActivityAt:       msg.Timestamp,
			AdditionalAttributes: json.RawMessage(fmt.Sprintf(`{"whatsapp_phone_number_id": "%s"}`, msg.PhoneNumberID)),
		}
		conv, err = s.conversationRepo.CreateTx(ctx, tx, newConv)
		if err != nil {
			return fmt.Errorf("failed to create conversation in tx: %w", err)
		}
	}

	// 7. Build and Persist Message in Tx (with ON CONFLICT DB-level idempotency)
	additionalAttrs := map[string]interface{}{
		"phone_number_id": msg.PhoneNumberID,
		"raw_payload":     json.RawMessage(rawBody),
	}
	if msg.ContextReplyID != "" {
		additionalAttrs["in_reply_to_external_id"] = msg.ContextReplyID
	}
	additionalAttrsBytes, _ := json.Marshal(additionalAttrs)

	externalSourceIDs := map[string]string{
		"whatsapp": msg.MetaMessageID,
	}
	externalSourceIDsBytes, _ := json.Marshal(externalSourceIDs)

	dbMsg := &model.Message{
		Content:              &msg.Content,
		AccountID:            accountID,
		InboxID:              inboxID,
		ConversationID:       conv.ID,
		MessageType:          model.MessageTypeIncoming,
		ContentType:          msg.ContentType,
		SenderType:           "Contact",
		SenderID:             ptrInt64(int64(contact.ID)),
		SourceID:             ptrString(msg.MetaMessageID),
		ExternalSourceIDs:    externalSourceIDsBytes,
		AdditionalAttributes: additionalAttrsBytes,
		Status:               model.MessageStatusSent,
		Private:              false,
		CreatedAt:            msg.Timestamp,
	}

	savedMsg, err := s.messageRepo.CreateTx(ctx, tx, dbMsg)
	if err != nil {
		return fmt.Errorf("failed to save message in tx: %w", err)
	}

	// 8. Persist Attachments in Tx if present
	attachmentsForEvent := make([]model.NormalizedAttachment, 0)
	for _, att := range msg.Attachments {
		metaBytes, _ := json.Marshal(att.Meta)
		dbAtt := &model.Attachment{
			FileType:        att.FileType,
			ExternalURL:     att.ExternalURL,
			CoordinatesLat:  att.CoordinatesLat,
			CoordinatesLong: att.CoordinatesLong,
			MessageID:       savedMsg.ID,
			AccountID:       accountID,
			FallbackTitle:   ptrString(att.FallbackTitle),
			Extension:       ptrString(att.Extension),
			Meta:            metaBytes,
		}
		savedAtt, err := s.messageRepo.CreateAttachmentTx(ctx, tx, dbAtt)
		if err == nil {
			attachmentsForEvent = append(attachmentsForEvent, model.NormalizedAttachment{
				ID:          savedAtt.ID,
				FileType:    attTypeToString(savedAtt.FileType),
				ExternalURL: savedAtt.ExternalURL,
				Title:       att.FallbackTitle,
				Extension:   att.Extension,
				Lat:         savedAtt.CoordinatesLat,
				Long:        savedAtt.CoordinatesLong,
			})
		}
	}

	// 9. Persist Raw Event to WhatsApp Gateway Audit Table in Tx
	now := time.Now().UTC()
	gatewayEvent := &model.WhatsAppGatewayEvent{
		AccountID:         &accountID,
		InboxID:           &inboxID,
		MessageID:         &savedMsg.ID,
		EventType:         "message_created",
		ExternalEventID:   msg.MetaMessageID,
		PhoneNumberID:     &msg.PhoneNumberID,
		RawPayload:        json.RawMessage(rawBody),
		ReceivedAt:        now,
		ProcessedAt:       &now,
	}
	_, _ = s.eventRepo.CreateEventTx(ctx, tx, gatewayEvent)

	// 10. Commit Atomic Transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit incoming message transaction: %w", err)
	}

	// 11. Publish Normalized Events to Redis Queue (Post-Commit)
	eventPayload := model.NormalizedWebhookEvent{
		Event:     queue.EventMessageCreated,
		AccountID: accountID,
		InboxID:   inboxID,
		Conversation: &model.NormalizedConversation{
			ID:             conv.ID,
			DisplayID:      conv.DisplayID,
			Status:         "open",
			UUID:           conv.UUID,
			LastActivityAt: conv.LastActivityAt,
			CreatedAt:      conv.CreatedAt,
		},
		Contact: &model.NormalizedContact{
			ID:          contact.ID,
			Name:        contact.Name,
			PhoneNumber: contact.PhoneNumber,
		},
		Message: &model.NormalizedMessage{
			ID:          savedMsg.ID,
			ExternalID:  msg.MetaMessageID,
			Type:        "incoming",
			ContentType: "text",
			Content:     msg.Content,
			Status:      "sent",
			CreatedAt:   savedMsg.CreatedAt,
			Attachments: attachmentsForEvent,
		},
		Timestamp: time.Now().UTC(),
	}

	if isNewConversation {
		convEvent := eventPayload
		convEvent.Event = queue.EventConversationCreated
		jobConv := queue.EventJob{
			ID:        uuid.New().String(),
			Type:      queue.EventConversationCreated,
			AccountID: accountID,
			InboxID:   inboxID,
			Event:     convEvent,
			CreatedAt: time.Now().UTC(),
		}
		_ = s.queue.Enqueue(ctx, queue.QueueIncomingMessages, jobConv)
	}

	jobMsg := queue.EventJob{
		ID:        uuid.New().String(),
		Type:      queue.EventMessageCreated,
		AccountID: accountID,
		InboxID:   inboxID,
		Event:     eventPayload,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.queue.Enqueue(ctx, queue.QueueIncomingMessages, jobMsg); err != nil {
		log.Printf("[WhatsAppService] Error enqueuing message job to Redis: %v", err)
	}

	log.Printf(`{"level":"info","event":"message_created","conversation_id":%d,"external_message_id":"%s","sender":"%s"}`,
		conv.ID, msg.MetaMessageID, msg.SenderPhone)

	return nil
}

func (s *whatsAppService) processStatusUpdate(ctx context.Context, status whatsapp.ParsedStatusUpdate, rawBody []byte) error {
	if status.MetaMessageID == "" {
		return nil
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for status: %w", err)
	}
	defer tx.Rollback(ctx)

	err = s.messageRepo.UpdateStatusBySourceIDTx(ctx, tx, status.MetaMessageID, status.Status)
	if err != nil {
		return fmt.Errorf("failed to update message status in db: %w", err)
	}

	// Audit event log for status update
	now := time.Now().UTC()
	gatewayEvent := &model.WhatsAppGatewayEvent{
		EventType:       "status_update_" + status.StatusStr,
		ExternalEventID: status.MetaMessageID + "_" + status.StatusStr,
		RawPayload:      json.RawMessage(rawBody),
		ReceivedAt:      now,
		ProcessedAt:     &now,
	}
	_, _ = s.eventRepo.CreateEventTx(ctx, tx, gatewayEvent)

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit status update transaction: %w", err)
	}

	// Enqueue status update event to Redis
	eventPayload := model.NormalizedWebhookEvent{
		Event: queue.EventMessageStatusUpdate,
		Message: &model.NormalizedMessage{
			ExternalID: status.MetaMessageID,
			Status:     status.StatusStr,
			CreatedAt:  status.Timestamp,
		},
		Timestamp: time.Now().UTC(),
	}

	job := queue.EventJob{
		ID:        uuid.New().String(),
		Type:      queue.EventMessageStatusUpdate,
		Event:     eventPayload,
		CreatedAt: time.Now().UTC(),
	}
	_ = s.queue.Enqueue(ctx, queue.QueueIncomingMessages, job)

	log.Printf(`{"level":"info","event":"message_status_updated","external_message_id":"%s","status":"%s"}`,
		status.MetaMessageID, status.StatusStr)

	return nil
}

func ptrString(s string) *string {
	return &s
}

func ptrInt64(i int64) *int64 {
	return &i
}

func attTypeToString(t int) string {
	switch t {
	case model.AttachmentTypeImage:
		return "image"
	case model.AttachmentTypeAudio:
		return "audio"
	case model.AttachmentTypeVideo:
		return "video"
	case model.AttachmentTypeFile:
		return "file"
	case model.AttachmentTypeLocation:
		return "location"
	case model.AttachmentTypeContact:
		return "contact"
	default:
		return "fallback"
	}
}

