package model

import "time"

type NormalizedWebhookEvent struct {
	Event        string                   `json:"event"`
	AccountID    int                      `json:"account_id"`
	InboxID      int                      `json:"inbox_id"`
	Conversation *NormalizedConversation `json:"conversation,omitempty"`
	Contact      *NormalizedContact      `json:"contact,omitempty"`
	Message      *NormalizedMessage      `json:"message,omitempty"`
	Timestamp    time.Time                `json:"timestamp"`
}

type NormalizedConversation struct {
	ID             int       `json:"id"`
	DisplayID      int       `json:"display_id"`
	Status         string    `json:"status"` // open, resolved, pending, snoozed
	UUID           string    `json:"uuid"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type NormalizedContact struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Identifier  string `json:"identifier,omitempty"`
}

type NormalizedMessage struct {
	ID          int                    `json:"id"`
	ExternalID  string                 `json:"external_id"`
	Type        string                 `json:"type"` // incoming, outgoing, activity, template
	ContentType string                 `json:"content_type"`
	Content     string                 `json:"content"`
	Status      string                 `json:"status"` // sent, delivered, read, failed
	CreatedAt   time.Time              `json:"created_at"`
	Attachments []NormalizedAttachment `json:"attachments,omitempty"`
}

type NormalizedAttachment struct {
	ID          int     `json:"id"`
	FileType    string  `json:"file_type"`
	ExternalURL string  `json:"external_url"`
	Title       string  `json:"title,omitempty"`
	Extension   string  `json:"extension,omitempty"`
	Lat         float64 `json:"latitude,omitempty"`
	Long        float64 `json:"longitude,omitempty"`
}
