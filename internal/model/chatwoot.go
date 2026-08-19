package model

import (
	"encoding/json"
	"time"
)

// Message Type constants matching Chatwoot
const (
	MessageTypeIncoming = 0
	MessageTypeOutgoing = 1
	MessageTypeActivity = 2
	MessageTypeTemplate = 3
)

// Message Status constants matching Chatwoot
const (
	MessageStatusSent      = 0
	MessageStatusDelivered = 1
	MessageStatusRead      = 2
	MessageStatusFailed    = 3
)

// Content Type constants matching Chatwoot
const (
	ContentTypeText       = 0
	ContentTypeInputText  = 1
	ContentTypeCards      = 5
	ContentTypeForm       = 6
	ContentTypeArticle    = 7
	ContentTypeIncomingEmail = 8
)

// Attachment File Type constants matching Chatwoot
const (
	AttachmentTypeImage    = 0
	AttachmentTypeAudio    = 1
	AttachmentTypeVideo    = 2
	AttachmentTypeFile     = 3
	AttachmentTypeLocation = 4
	AttachmentTypeFallback = 5
	AttachmentTypeContact  = 7
)

// Conversation Status constants matching Chatwoot
const (
	ConversationStatusOpen     = 0
	ConversationStatusResolved = 1
	ConversationStatusPending  = 2
	ConversationStatusSnoozed  = 3
)

type Account struct {
	ID        int             `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Locale    int             `json:"locale" db:"locale"`
	Settings  json.RawMessage `json:"settings" db:"settings"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type ChannelWhatsapp struct {
	ID                      int64           `json:"id" db:"id"`
	AccountID               int             `json:"account_id" db:"account_id"`
	PhoneNumber             string          `json:"phone_number" db:"phone_number"`
	Provider                string          `json:"provider" db:"provider"`
	ProviderConfig          json.RawMessage `json:"provider_config" db:"provider_config"`
	BusinessManagementToken *string         `json:"business_management_token,omitempty" db:"business_management_token"`
	CreatedAt               time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at" db:"updated_at"`
}

type Inbox struct {
	ID                        int       `json:"id" db:"id"`
	ChannelID                 int       `json:"channel_id" db:"channel_id"`
	AccountID                 int       `json:"account_id" db:"account_id"`
	Name                      string    `json:"name" db:"name"`
	ChannelType               string    `json:"channel_type" db:"channel_type"`
	LockToSingleConversation bool      `json:"lock_to_single_conversation" db:"lock_to_single_conversation"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

type Contact struct {
	ID                   int             `json:"id" db:"id"`
	Name                 string          `json:"name" db:"name"`
	Email                *string         `json:"email,omitempty" db:"email"`
	PhoneNumber          string          `json:"phone_number" db:"phone_number"`
	AccountID            int             `json:"account_id" db:"account_id"`
	Identifier           *string         `json:"identifier,omitempty" db:"identifier"`
	AdditionalAttributes json.RawMessage `json:"additional_attributes" db:"additional_attributes"`
	CustomAttributes     json.RawMessage `json:"custom_attributes" db:"custom_attributes"`
	LastActivityAt       *time.Time      `json:"last_activity_at,omitempty" db:"last_activity_at"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`
}

type ContactInbox struct {
	ID        int64     `json:"id" db:"id"`
	ContactID int64     `json:"contact_id" db:"contact_id"`
	InboxID   int64     `json:"inbox_id" db:"inbox_id"`
	SourceID  string    `json:"source_id" db:"source_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Conversation struct {
	ID                   int             `json:"id" db:"id"`
	AccountID            int             `json:"account_id" db:"account_id"`
	InboxID              int             `json:"inbox_id" db:"inbox_id"`
	ContactID            int64           `json:"contact_id" db:"contact_id"`
	ContactInboxID       int64           `json:"contact_inbox_id" db:"contact_inbox_id"`
	Status               int             `json:"status" db:"status"`
	DisplayID            int             `json:"display_id" db:"display_id"`
	UUID                 string          `json:"uuid" db:"uuid"`
	Identifier           *string         `json:"identifier,omitempty" db:"identifier"`
	LastActivityAt       time.Time       `json:"last_activity_at" db:"last_activity_at"`
	AdditionalAttributes json.RawMessage `json:"additional_attributes" db:"additional_attributes"`
	CustomAttributes     json.RawMessage `json:"custom_attributes" db:"custom_attributes"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at" db:"updated_at"`

	// Populated for API responses
	Contact     *Contact     `json:"contact,omitempty"`
	LastMessage *Message     `json:"last_message,omitempty"`
	UnreadCount int          `json:"unread_count,omitempty"`
}

type Message struct {
	ID                      int             `json:"id" db:"id"`
	Content                 *string         `json:"content" db:"content"`
	AccountID               int             `json:"account_id" db:"account_id"`
	InboxID                 int             `json:"inbox_id" db:"inbox_id"`
	ConversationID          int             `json:"conversation_id" db:"conversation_id"`
	MessageType             int             `json:"message_type" db:"message_type"`
	ContentType             int             `json:"content_type" db:"content_type"`
	ContentAttributes       json.RawMessage `json:"content_attributes" db:"content_attributes"`
	SenderType              string          `json:"sender_type" db:"sender_type"`
	SenderID                *int64          `json:"sender_id,omitempty" db:"sender_id"`
	SourceID                *string         `json:"source_id,omitempty" db:"source_id"`
	ExternalSourceIDs       json.RawMessage `json:"external_source_ids" db:"external_source_ids"`
	AdditionalAttributes    json.RawMessage `json:"additional_attributes" db:"additional_attributes"`
	ProcessedMessageContent *string         `json:"processed_message_content,omitempty" db:"processed_message_content"`
	Status                  int             `json:"status" db:"status"`
	Private                 bool            `json:"private" db:"private"`
	CreatedAt               time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at" db:"updated_at"`

	// Populated for API responses
	Attachments []Attachment `json:"attachments,omitempty"`
	Sender      *Contact     `json:"sender,omitempty"`
}

type Attachment struct {
	ID              int             `json:"id" db:"id"`
	FileType        int             `json:"file_type" db:"file_type"`
	ExternalURL     string          `json:"external_url" db:"external_url"`
	CoordinatesLat  float64         `json:"coordinates_lat" db:"coordinates_lat"`
	CoordinatesLong float64         `json:"coordinates_long" db:"coordinates_long"`
	MessageID       int             `json:"message_id" db:"message_id"`
	AccountID       int             `json:"account_id" db:"account_id"`
	FallbackTitle   *string         `json:"fallback_title,omitempty" db:"fallback_title"`
	Extension       *string         `json:"extension,omitempty" db:"extension"`
	Meta            json.RawMessage `json:"meta" db:"meta"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type Webhook struct {
	ID            int64           `json:"id" db:"id"`
	AccountID     *int            `json:"account_id,omitempty" db:"account_id"`
	InboxID       *int            `json:"inbox_id,omitempty" db:"inbox_id"`
	URL           string          `json:"url" db:"url"`
	WebhookType   int             `json:"webhook_type" db:"webhook_type"`
	Subscriptions json.RawMessage `json:"subscriptions" db:"subscriptions"`
	Name          *string         `json:"name,omitempty" db:"name"`
	Secret        *string         `json:"secret,omitempty" db:"secret"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type WebhookDeliveryAttempt struct {
	ID            int64           `json:"id" db:"id"`
	WebhookID     int64           `json:"webhook_id" db:"webhook_id"`
	Event         string          `json:"event" db:"event"`
	Payload       json.RawMessage `json:"payload" db:"payload"`
	URL           string          `json:"url" db:"url"`
	Attempts      int             `json:"attempts" db:"attempts"`
	Status        string          `json:"status" db:"status"` // pending, success, failed, dead_letter
	LastError     *string         `json:"last_error,omitempty" db:"last_error"`
	ResponseCode  *int            `json:"response_code,omitempty" db:"response_code"`
	LastAttemptAt *time.Time      `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
	NextAttemptAt *time.Time      `json:"next_attempt_at,omitempty" db:"next_attempt_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

type WhatsAppGatewayEvent struct {
	ID                 int64           `json:"id" db:"id"`
	AccountID          *int            `json:"account_id,omitempty" db:"account_id"`
	InboxID            *int            `json:"inbox_id,omitempty" db:"inbox_id"`
	MessageID          *int            `json:"message_id,omitempty" db:"message_id"`
	EventType          string          `json:"event_type" db:"event_type"`
	ExternalEventID    string          `json:"external_event_id" db:"external_event_id"`
	PhoneNumberID      *string         `json:"phone_number_id,omitempty" db:"phone_number_id"`
	BusinessAccountID  *string         `json:"business_account_id,omitempty" db:"business_account_id"`
	RawPayload         json.RawMessage `json:"raw_payload" db:"raw_payload"`
	ReceivedAt         time.Time       `json:"received_at" db:"received_at"`
	ProcessedAt        *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
}
