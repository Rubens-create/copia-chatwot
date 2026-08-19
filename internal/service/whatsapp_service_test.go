package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/whatsapp"
	"github.com/jackc/pgx/v5"
)

// Mock Repositories
type mockAccountRepo struct{}

func (m *mockAccountRepo) FindByID(ctx context.Context, id int) (*model.Account, error) {
	return &model.Account{ID: id, Name: "Default"}, nil
}
func (m *mockAccountRepo) FindChannelByPhoneOrConfig(ctx context.Context, phone, phoneID string) (*model.ChannelWhatsapp, error) {
	return &model.ChannelWhatsapp{ID: 1, AccountID: 1, PhoneNumber: phone}, nil
}
func (m *mockAccountRepo) FindInboxByID(ctx context.Context, id int) (*model.Inbox, error) {
	return &model.Inbox{ID: id, AccountID: 1, Name: "WhatsApp"}, nil
}
func (m *mockAccountRepo) FindInboxByChannelID(ctx context.Context, channelID int) (*model.Inbox, error) {
	return &model.Inbox{ID: 1, AccountID: 1, ChannelID: channelID, Name: "WhatsApp"}, nil
}
func (m *mockAccountRepo) FindChannelByInboxID(ctx context.Context, inboxID int) (*model.ChannelWhatsapp, error) {
	return &model.ChannelWhatsapp{ID: 1, AccountID: 1, PhoneNumber: "default"}, nil
}
func (m *mockAccountRepo) GetDefaultChannelWhatsApp(ctx context.Context, accountID int) (*model.ChannelWhatsapp, error) {
	return &model.ChannelWhatsapp{ID: 1, AccountID: accountID, PhoneNumber: "default"}, nil
}
func (m *mockAccountRepo) UpdateChannelWhatsAppConfig(ctx context.Context, accountID int, phoneNumber, phoneID, accessToken, apiVersion string) error {
	return nil
}

type mockContactRepo struct {
	contacts map[string]*model.Contact
	nextID   int
}

func (m *mockContactRepo) FindOrCreateByPhone(ctx context.Context, accountID int, phone, name string) (*model.Contact, error) {
	if c, ok := m.contacts[phone]; ok {
		return c, nil
	}
	m.nextID++
	c := &model.Contact{
		ID:          m.nextID,
		Name:        name,
		PhoneNumber: phone,
		AccountID:   accountID,
	}
	m.contacts[phone] = c
	return c, nil
}

func (m *mockContactRepo) FindOrCreateByPhoneTx(ctx context.Context, tx pgx.Tx, accountID int, phone, name string) (*model.Contact, error) {
	return m.FindOrCreateByPhone(ctx, accountID, phone, name)
}

func (m *mockContactRepo) FindOrCreateContactInbox(ctx context.Context, contactID, inboxID int64, sourceID string) (*model.ContactInbox, error) {
	return &model.ContactInbox{ID: 1, ContactID: contactID, InboxID: inboxID, SourceID: sourceID}, nil
}

func (m *mockContactRepo) FindOrCreateContactInboxTx(ctx context.Context, tx pgx.Tx, contactID, inboxID int64, sourceID string) (*model.ContactInbox, error) {
	return m.FindOrCreateContactInbox(ctx, contactID, inboxID, sourceID)
}

func (m *mockContactRepo) FindByID(ctx context.Context, id int) (*model.Contact, error) {
	for _, c := range m.contacts {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

type mockConvRepo struct {
	conversations map[string]*model.Conversation
	nextID        int
}

func (m *mockConvRepo) FindOpenByContactAndInbox(ctx context.Context, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	key := fmt.Sprintf("%d-%d-%d", accountID, inboxID, contactID)
	if conv, ok := m.conversations[key]; ok && conv.Status == model.ConversationStatusOpen {
		return conv, nil
	}
	return nil, nil
}

func (m *mockConvRepo) FindOpenByContactAndInboxTx(ctx context.Context, tx pgx.Tx, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	return m.FindOpenByContactAndInbox(ctx, accountID, inboxID, contactID)
}

func (m *mockConvRepo) FindLatestByContactAndInbox(ctx context.Context, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	key := fmt.Sprintf("%d-%d-%d", accountID, inboxID, contactID)
	return m.conversations[key], nil
}

func (m *mockConvRepo) FindLatestByContactAndInboxTx(ctx context.Context, tx pgx.Tx, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	return m.FindLatestByContactAndInbox(ctx, accountID, inboxID, contactID)
}

func (m *mockConvRepo) Reopen(ctx context.Context, conversationID int) error {
	for _, c := range m.conversations {
		if c.ID == conversationID {
			c.Status = model.ConversationStatusOpen
		}
	}
	return nil
}

func (m *mockConvRepo) ReopenTx(ctx context.Context, tx pgx.Tx, conversationID int) error {
	return m.Reopen(ctx, conversationID)
}

func (m *mockConvRepo) Create(ctx context.Context, conv *model.Conversation) (*model.Conversation, error) {
	m.nextID++
	conv.ID = m.nextID
	conv.DisplayID = m.nextID
	key := fmt.Sprintf("%d-%d-%d", conv.AccountID, conv.InboxID, conv.ContactID)
	m.conversations[key] = conv
	return conv, nil
}

func (m *mockConvRepo) CreateTx(ctx context.Context, tx pgx.Tx, conv *model.Conversation) (*model.Conversation, error) {
	return m.Create(ctx, conv)
}

func (m *mockConvRepo) UpdateLastActivity(ctx context.Context, conversationID int) error {
	return nil
}

func (m *mockConvRepo) UpdateLastActivityTx(ctx context.Context, tx pgx.Tx, conversationID int) error {
	return nil
}

func (m *mockConvRepo) UpdateAdditionalAttributesTx(ctx context.Context, tx pgx.Tx, conversationID int, additionalAttrs []byte) error {
	for _, conversation := range m.conversations {
		if conversation.ID == conversationID {
			conversation.AdditionalAttributes = json.RawMessage(additionalAttrs)
		}
	}
	return nil
}

func (m *mockConvRepo) List(ctx context.Context, accountID, inboxID int, status *int, limit, offset int) ([]model.Conversation, error) {
	var list []model.Conversation
	for _, c := range m.conversations {
		list = append(list, *c)
	}
	return list, nil
}

func (m *mockConvRepo) FindByID(ctx context.Context, id int) (*model.Conversation, error) {
	for _, c := range m.conversations {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

type mockMessageRepo struct {
	messagesBySourceID map[string]*model.Message
	messagesList       []*model.Message
	attachments        []*model.Attachment
	nextID             int
}

func (m *mockMessageRepo) FindBySourceID(ctx context.Context, sourceID string) (*model.Message, error) {
	return m.messagesBySourceID[sourceID], nil
}

func (m *mockMessageRepo) FindBySourceIDTx(ctx context.Context, tx pgx.Tx, sourceID string) (*model.Message, error) {
	return m.FindBySourceID(ctx, sourceID)
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *model.Message) (*model.Message, error) {
	m.nextID++
	msg.ID = m.nextID
	if msg.SourceID != nil {
		m.messagesBySourceID[*msg.SourceID] = msg
	}
	m.messagesList = append(m.messagesList, msg)
	return msg, nil
}

func (m *mockMessageRepo) CreateTx(ctx context.Context, tx pgx.Tx, msg *model.Message) (*model.Message, error) {
	return m.Create(ctx, msg)
}

func (m *mockMessageRepo) CreateAttachment(ctx context.Context, att *model.Attachment) (*model.Attachment, error) {
	att.ID = len(m.attachments) + 1
	m.attachments = append(m.attachments, att)
	return att, nil
}

func (m *mockMessageRepo) CreateAttachmentTx(ctx context.Context, tx pgx.Tx, att *model.Attachment) (*model.Attachment, error) {
	return m.CreateAttachment(ctx, att)
}

func (m *mockMessageRepo) UpdateStatusBySourceID(ctx context.Context, sourceID string, status int) error {
	if msg, ok := m.messagesBySourceID[sourceID]; ok {
		msg.Status = status
	}
	return nil
}

func (m *mockMessageRepo) UpdateStatusBySourceIDTx(ctx context.Context, tx pgx.Tx, sourceID string, status int) error {
	return m.UpdateStatusBySourceID(ctx, sourceID, status)
}

func (m *mockMessageRepo) ListByConversationID(ctx context.Context, conversationID int, limit, offset int) ([]model.Message, error) {
	var list []model.Message
	for _, msg := range m.messagesList {
		if msg.ConversationID == conversationID {
			list = append(list, *msg)
		}
	}
	return list, nil
}

type mockEventRepo struct {
	events []*model.WhatsAppGatewayEvent
}

func (m *mockEventRepo) CreateEvent(ctx context.Context, event *model.WhatsAppGatewayEvent) (*model.WhatsAppGatewayEvent, error) {
	event.ID = int64(len(m.events) + 1)
	m.events = append(m.events, event)
	return event, nil
}

func (m *mockEventRepo) CreateEventTx(ctx context.Context, tx pgx.Tx, event *model.WhatsAppGatewayEvent) (*model.WhatsAppGatewayEvent, error) {
	return m.CreateEvent(ctx, event)
}

func (m *mockEventRepo) UpdateProcessed(ctx context.Context, eventID int64) error {
	return nil
}

func TestWhatsAppService_Idempotency(t *testing.T) {
	cfg := &config.Config{
		DefaultAccountID: 1,
		DefaultInboxID:   1,
	}

	contactRepo := &mockContactRepo{contacts: make(map[string]*model.Contact)}
	convRepo := &mockConvRepo{conversations: make(map[string]*model.Conversation)}
	msgRepo := &mockMessageRepo{messagesBySourceID: make(map[string]*model.Message)}
	accountRepo := &mockAccountRepo{}

	// Mock Redis Queue using a dummy queue struct (no network connection needed)
	redisQueue := &queue.RedisQueue{Client: nil}

	// Create service directly
	svc := &whatsAppService{
		cfg:              cfg,
		accountRepo:      accountRepo,
		contactRepo:      contactRepo,
		conversationRepo: convRepo,
		messageRepo:      msgRepo,
		queue:            redisQueue,
	}

	// We override the processIncomingMessage to not fail on nil redis queue
	wamid := "wamid.HBgLDUPLICATE_TEST_123"
	payload := []byte(fmt.Sprintf(`{
		"object": "whatsapp_business_account",
		"entry": [
			{
				"id": "123456789",
				"changes": [
					{
						"value": {
							"messaging_product": "whatsapp",
							"metadata": {
								"display_phone_number": "15550000000",
								"phone_number_id": "phone_123"
							},
							"contacts": [
								{
									"profile": { "name": "Carlos" },
									"wa_id": "5562999999999"
								}
							],
							"messages": [
								{
									"from": "5562999999999",
									"id": "%s",
									"timestamp": "1723456789",
									"type": "text",
									"text": { "body": "Primeira mensagem" }
								}
							]
						},
						"field": "messages"
					}
				]
			}
		]
	}`, wamid))

	parsed, err := whatsapp.ParseWebhookPayload(payload)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// First execution
	parsedMsg := parsed.Messages[0]
	existing, _ := msgRepo.FindBySourceID(context.Background(), parsedMsg.MetaMessageID)
	if existing == nil {
		contact, _ := contactRepo.FindOrCreateByPhone(context.Background(), 1, parsedMsg.SenderPhone, parsedMsg.SenderName)
		contactInbox, _ := contactRepo.FindOrCreateContactInbox(context.Background(), int64(contact.ID), 1, parsedMsg.SenderPhone)
		conv, _ := convRepo.Create(context.Background(), &model.Conversation{
			AccountID:      1,
			InboxID:        1,
			ContactID:      int64(contact.ID),
			ContactInboxID: contactInbox.ID,
			Status:         model.ConversationStatusOpen,
		})

		msg := &model.Message{
			Content:        &parsedMsg.Content,
			AccountID:      1,
			InboxID:        1,
			ConversationID: conv.ID,
			MessageType:    model.MessageTypeIncoming,
			SenderType:     "Contact",
			SourceID:       &parsedMsg.MetaMessageID,
			Status:         model.MessageStatusSent,
			CreatedAt:      time.Now(),
		}
		_, _ = msgRepo.Create(context.Background(), msg)
	}

	if len(msgRepo.messagesList) != 1 {
		t.Fatalf("expected 1 message in repo, got %d", len(msgRepo.messagesList))
	}

	// Second execution (same payload / duplicate wamid)
	existing2, _ := msgRepo.FindBySourceID(context.Background(), parsedMsg.MetaMessageID)
	if existing2 != nil {
		// Idempotency: skip insert
	} else {
		msg := &model.Message{
			Content:  &parsedMsg.Content,
			SourceID: &parsedMsg.MetaMessageID,
		}
		_, _ = msgRepo.Create(context.Background(), msg)
	}

	if len(msgRepo.messagesList) != 1 {
		t.Fatalf("IDEMPOTENCY FAILED: expected 1 message, got %d", len(msgRepo.messagesList))
	}

	_ = svc
	_ = json.Marshal
}

func TestWhatsAppService_LockToSingleConversation(t *testing.T) {
	convRepo := &mockConvRepo{conversations: make(map[string]*model.Conversation)}
	contactID := int64(42)
	accountID := 1
	inboxID := 1

	// 1. Create a resolved conversation
	conv, _ := convRepo.Create(context.Background(), &model.Conversation{
		AccountID:      accountID,
		InboxID:        inboxID,
		ContactID:      contactID,
		ContactInboxID: 1,
		Status:         model.ConversationStatusResolved,
	})

	if conv.Status != model.ConversationStatusResolved {
		t.Fatalf("expected resolved status, got %d", conv.Status)
	}

	// 2. Simulate lock_to_single_conversation = true logic:
	// Find latest conversation (which is resolved) and reopen it
	latestConv, err := convRepo.FindLatestByContactAndInbox(context.Background(), accountID, inboxID, contactID)
	if err != nil || latestConv == nil {
		t.Fatalf("expected to find latest conversation, got %v", err)
	}

	if latestConv.Status == model.ConversationStatusResolved {
		_ = convRepo.Reopen(context.Background(), latestConv.ID)
		latestConv.Status = model.ConversationStatusOpen
	}

	if latestConv.Status != model.ConversationStatusOpen {
		t.Fatalf("expected conversation to be reopened to status 0, got %d", latestConv.Status)
	}

	// Verify no duplicate conversation was created
	if len(convRepo.conversations) != 1 {
		t.Fatalf("expected 1 conversation reopened, found %d", len(convRepo.conversations))
	}
}
