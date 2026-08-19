package queue

import (
	"encoding/json"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
)

const (
	QueueIncomingMessages = "queue:incoming_messages"
	QueueOutgoingWebhooks = "queue:outgoing_webhooks"
	QueueRetry            = "queue:retry_webhooks"
	QueueDeadLetter       = "queue:dead_letter_webhooks"
)

const (
	EventMessageCreated      = "message_created"
	EventMessageUpdated      = "message_updated"
	EventMessageStatusUpdate = "message_status_updated"
	EventConversationCreated = "conversation_created"
	EventConversationUpdated = "conversation_updated"
)

type EventJob struct {
	ID        string                       `json:"id"`
	Type      string                       `json:"type"` // EventMessageCreated, etc.
	AccountID int                          `json:"account_id"`
	InboxID   int                          `json:"inbox_id"`
	Event     model.NormalizedWebhookEvent `json:"event"`
	CreatedAt time.Time                    `json:"created_at"`
}

type WebhookDeliveryJob struct {
	ID          string          `json:"id"`
	WebhookID   int64           `json:"webhook_id"`
	URL         string          `json:"url"`
	Secret      string          `json:"secret,omitempty"`
	Event       string          `json:"event"`
	Payload     json.RawMessage `json:"payload"`
	Attempt     int             `json:"attempt"`
	MaxAttempts int             `json:"max_attempts"`
	CreatedAt   time.Time       `json:"created_at"`
}
