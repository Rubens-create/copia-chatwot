package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/repository"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/whatsapp"
	"github.com/google/uuid"
)

type AttachmentParam struct {
	FileType        int     `json:"file_type"` // 0=image, 1=audio, 2=video, 3=file, 4=location, 7=contact
	ExternalURL     string  `json:"external_url"`
	DataURL         string  `json:"data_url,omitempty"`
	FallbackTitle   string  `json:"fallback_title,omitempty"`
	Extension       string  `json:"extension,omitempty"`
	CoordinatesLat  float64 `json:"coordinates_lat,omitempty"`
	CoordinatesLong float64 `json:"coordinates_long,omitempty"`
}

type TemplateParam struct {
	Name       string        `json:"name"`
	Language   string        `json:"language"`
	Components []interface{} `json:"components"`
}

type SendMessageParams struct {
	Content            string            `json:"content"`
	Attachments        []AttachmentParam `json:"attachments,omitempty"`
	Template           *TemplateParam    `json:"template,omitempty"`
	TemplateName       string            `json:"template_name,omitempty"`
	TemplateLanguage   string            `json:"template_language,omitempty"`
	TemplateComponents []interface{}     `json:"template_components,omitempty"`
	Private            bool              `json:"private"`
}

type ConversationService interface {
	ListConversations(ctx context.Context, accountID, inboxID int, status *int, limit, offset int) ([]model.Conversation, error)
	GetConversation(ctx context.Context, id int) (*model.Conversation, error)
	GetMessages(ctx context.Context, conversationID int, limit, offset int) ([]model.Message, error)
	SendMessage(ctx context.Context, conversationID int, params SendMessageParams) (*model.Message, error)
}

type conversationService struct {
	cfg              *config.Config
	accountRepo      repository.AccountRepository
	contactRepo      repository.ContactRepository
	conversationRepo repository.ConversationRepository
	messageRepo      repository.MessageRepository
	waClient         whatsapp.WhatsAppClient
	queue            *queue.RedisQueue
}

func NewConversationService(
	cfg *config.Config,
	accountRepo repository.AccountRepository,
	contactRepo repository.ContactRepository,
	conversationRepo repository.ConversationRepository,
	messageRepo repository.MessageRepository,
	waClient whatsapp.WhatsAppClient,
	queue *queue.RedisQueue,
) ConversationService {
	return &conversationService{
		cfg:              cfg,
		accountRepo:      accountRepo,
		contactRepo:      contactRepo,
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		waClient:         waClient,
		queue:            queue,
	}
}

func (s *conversationService) ListConversations(ctx context.Context, accountID, inboxID int, status *int, limit, offset int) ([]model.Conversation, error) {
	if accountID == 0 {
		accountID = s.cfg.DefaultAccountID
	}
	return s.conversationRepo.List(ctx, accountID, inboxID, status, limit, offset)
}

func (s *conversationService) GetConversation(ctx context.Context, id int) (*model.Conversation, error) {
	return s.conversationRepo.FindByID(ctx, id)
}

func (s *conversationService) GetMessages(ctx context.Context, conversationID int, limit, offset int) ([]model.Message, error) {
	return s.messageRepo.ListByConversationID(ctx, conversationID, limit, offset)
}

func (s *conversationService) SendMessage(ctx context.Context, conversationID int, params SendMessageParams) (*model.Message, error) {
	conv, err := s.conversationRepo.FindByID(ctx, conversationID)
	if err != nil || conv == nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	contact, err := s.contactRepo.FindByID(ctx, int(conv.ContactID))
	if err != nil || contact == nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}

	phoneID := "default"
	var addAttrs map[string]interface{}
	if len(conv.AdditionalAttributes) > 0 {
		_ = json.Unmarshal(conv.AdditionalAttributes, &addAttrs)
		if pid, ok := addAttrs["whatsapp_phone_number_id"].(string); ok && pid != "" {
			phoneID = pid
		}
	}

	var sourceID string
	contentType := model.ContentTypeText
	var contentAttributes json.RawMessage

	// 1. Template / HSM Message Handling
	templateName := params.TemplateName
	templateLang := params.TemplateLanguage
	var templateComponents []interface{} = params.TemplateComponents

	if params.Template != nil {
		templateName = params.Template.Name
		if params.Template.Language != "" {
			templateLang = params.Template.Language
		}
		if len(params.Template.Components) > 0 {
			templateComponents = params.Template.Components
		}
	}

	if templateName != "" {
		if templateLang == "" {
			templateLang = "pt_BR"
		}
		contentType = 1 // Template
		attrMap := map[string]interface{}{
			"template_name":       templateName,
			"template_language":   templateLang,
			"template_components": templateComponents,
		}
		contentAttributes, _ = json.Marshal(attrMap)

		if params.Content == "" {
			params.Content = fmt.Sprintf("[Template: %s]", templateName)
		}

		if s.waClient != nil && s.cfg.MetaAccessToken != "" {
			resp, err := s.waClient.SendTemplate(ctx, phoneID, contact.PhoneNumber, templateName, templateLang, templateComponents)
			if err != nil {
				log.Printf("[ConversationService] Warning: Meta API send template failed (%v). Persisting locally.", err)
			} else if resp != nil && len(resp.Messages) > 0 {
				sourceID = resp.Messages[0].ID
			}
		}
	} else if len(params.Attachments) > 0 {
		// 2. Attachments / Media Message Handling
		firstAtt := params.Attachments[0]
		url := firstAtt.ExternalURL
		if url == "" {
			url = firstAtt.DataURL
		}

		if s.waClient != nil && s.cfg.MetaAccessToken != "" && url != "" {
			var resp *whatsapp.SendMessageResponse
			var sendErr error

			switch firstAtt.FileType {
			case 0: // Image
				resp, sendErr = s.waClient.SendImage(ctx, phoneID, contact.PhoneNumber, url, params.Content)
			case 1: // Audio
				resp, sendErr = s.waClient.SendAudio(ctx, phoneID, contact.PhoneNumber, url)
			case 2: // Video
				resp, sendErr = s.waClient.SendVideo(ctx, phoneID, contact.PhoneNumber, url, params.Content)
			case 3: // File/Document
				resp, sendErr = s.waClient.SendDocument(ctx, phoneID, contact.PhoneNumber, url, firstAtt.FallbackTitle, params.Content)
			case 4: // Location
				resp, sendErr = s.waClient.SendLocation(ctx, phoneID, contact.PhoneNumber, firstAtt.CoordinatesLat, firstAtt.CoordinatesLong, firstAtt.FallbackTitle, "")
			default:
				resp, sendErr = s.waClient.SendDocument(ctx, phoneID, contact.PhoneNumber, url, firstAtt.FallbackTitle, params.Content)
			}

			if sendErr != nil {
				log.Printf("[ConversationService] Warning: Meta API media send failed (%v). Persisting locally.", sendErr)
			} else if resp != nil && len(resp.Messages) > 0 {
				sourceID = resp.Messages[0].ID
			}
		}
	} else {
		// 3. Plain Text Message
		if s.waClient != nil && s.cfg.MetaAccessToken != "" && params.Content != "" {
			resp, err := s.waClient.SendText(ctx, phoneID, contact.PhoneNumber, params.Content)
			if err != nil {
				log.Printf("[ConversationService] Warning: Meta API send text failed (%v). Persisting locally.", err)
			} else if resp != nil && len(resp.Messages) > 0 {
				sourceID = resp.Messages[0].ID
			}
		}
	}

	if sourceID == "" {
		sourceID = fmt.Sprintf("out.wamid.%s", uuid.New().String())
	}

	now := time.Now().UTC()
	externalSourceIDs, _ := json.Marshal(map[string]string{"whatsapp": sourceID})

	var contentPtr *string
	if params.Content != "" {
		contentPtr = &params.Content
	}

	msg := &model.Message{
		Content:           contentPtr,
		AccountID:         conv.AccountID,
		InboxID:           conv.InboxID,
		ConversationID:    conv.ID,
		MessageType:       model.MessageTypeOutgoing,
		ContentType:       contentType,
		ContentAttributes: contentAttributes,
		SenderType:        "User",
		SourceID:          &sourceID,
		ExternalSourceIDs: externalSourceIDs,
		Status:            model.MessageStatusSent,
		Private:           params.Private,
		CreatedAt:         now,
	}

	created, err := s.messageRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to persist outgoing message: %w", err)
	}

	// Persist Attachments in database
	for _, att := range params.Attachments {
		extURL := att.ExternalURL
		if extURL == "" {
			extURL = att.DataURL
		}
		var fallbackTitlePtr *string
		if att.FallbackTitle != "" {
			fallbackTitlePtr = &att.FallbackTitle
		}
		var extPtr *string
		if att.Extension != "" {
			extPtr = &att.Extension
		}

		dbAtt := &model.Attachment{
			AccountID:       conv.AccountID,
			MessageID:       created.ID,
			FileType:        att.FileType,
			ExternalURL:     extURL,
			CoordinatesLat:  att.CoordinatesLat,
			CoordinatesLong: att.CoordinatesLong,
			FallbackTitle:   fallbackTitlePtr,
			Extension:       extPtr,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		savedAtt, err := s.messageRepo.CreateAttachment(ctx, dbAtt)
		if err == nil && savedAtt != nil {
			created.Attachments = append(created.Attachments, *savedAtt)
		}
	}

	_ = s.conversationRepo.UpdateLastActivity(ctx, conv.ID)

	// Publish event to Redis for outbound webhooks
	eventPayload := model.NormalizedWebhookEvent{
		Event:     queue.EventMessageCreated,
		AccountID: conv.AccountID,
		InboxID:   conv.InboxID,
		Conversation: &model.NormalizedConversation{
			ID:             conv.ID,
			DisplayID:      conv.DisplayID,
			Status:         "open",
			UUID:           conv.UUID,
			LastActivityAt: now,
			CreatedAt:      conv.CreatedAt,
		},
		Contact: &model.NormalizedContact{
			ID:          contact.ID,
			Name:        contact.Name,
			PhoneNumber: contact.PhoneNumber,
		},
		Message: &model.NormalizedMessage{
			ID:          created.ID,
			ExternalID:  sourceID,
			Type:        "outgoing",
			ContentType: "text",
			Content:     params.Content,
			Status:      "sent",
			CreatedAt:   now,
		},
		Timestamp: now,
	}

	job := queue.EventJob{
		ID:        uuid.New().String(),
		Type:      queue.EventMessageCreated,
		AccountID: conv.AccountID,
		InboxID:   conv.InboxID,
		Event:     eventPayload,
		CreatedAt: now,
	}
	_ = s.queue.Enqueue(ctx, queue.QueueIncomingMessages, job)

	return created, nil
}
