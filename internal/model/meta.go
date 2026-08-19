package model

import "encoding/json"

type MetaWebhookPayload struct {
	Object string          `json:"object"`
	Entry  []MetaEntry     `json:"entry"`
}

type MetaEntry struct {
	ID      string       `json:"id"`
	Changes []MetaChange `json:"changes"`
}

type MetaChange struct {
	Value MetaValue `json:"value"`
	Field string    `json:"field"`
}

type MetaValue struct {
	MessagingProduct string            `json:"messaging_product"`
	Metadata         MetaMetadata      `json:"metadata"`
	Contacts         []MetaContact     `json:"contacts,omitempty"`
	Messages         []MetaMessage     `json:"messages,omitempty"`
	Statuses         []MetaStatus      `json:"statuses,omitempty"`
	Errors           []json.RawMessage `json:"errors,omitempty"`
}

type MetaMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type MetaContact struct {
	Profile MetaProfile `json:"profile"`
	WaID    string      `json:"wa_id"`
}

type MetaProfile struct {
	Name string `json:"name"`
}

type MetaMessage struct {
	From        string             `json:"from"`
	ID          string             `json:"id"` // wamid.HBg...
	Timestamp   string             `json:"timestamp"`
	Type        string             `json:"type"` // text, image, audio, video, document, location, contacts, interactive, sticker
	Text        *MetaText          `json:"text,omitempty"`
	Image       *MetaMedia         `json:"image,omitempty"`
	Audio       *MetaMedia         `json:"audio,omitempty"`
	Video       *MetaMedia         `json:"video,omitempty"`
	Document    *MetaMedia         `json:"document,omitempty"`
	Sticker     *MetaMedia         `json:"sticker,omitempty"`
	Location    *MetaLocation      `json:"location,omitempty"`
	Contacts    []MetaSharedContact `json:"contacts,omitempty"`
	Interactive *MetaInteractive   `json:"interactive,omitempty"`
	Context     *MetaContext       `json:"context,omitempty"`
	Errors      []json.RawMessage  `json:"errors,omitempty"`
}

type MetaText struct {
	Body string `json:"body"`
}

type MetaMedia struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	Sha256   string `json:"sha256,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
	Voice    bool   `json:"voice,omitempty"`
}

type MetaLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type MetaSharedContact struct {
	Name   MetaContactName    `json:"name"`
	Phones []MetaContactPhone `json:"phones,omitempty"`
	Emails []MetaContactEmail `json:"emails,omitempty"`
}

type MetaContactName struct {
	FormattedName string `json:"formatted_name"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
}

type MetaContactPhone struct {
	Phone string `json:"phone"`
	Type  string `json:"type,omitempty"`
	WaID  string `json:"wa_id,omitempty"`
}

type MetaContactEmail struct {
	Email string `json:"email"`
	Type  string `json:"type,omitempty"`
}

type MetaInteractive struct {
	Type        string               `json:"type"`
	ButtonReply *MetaInteractiveReply `json:"button_reply,omitempty"`
	ListReply   *MetaInteractiveReply `json:"list_reply,omitempty"`
}

type MetaInteractiveReply struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type MetaContext struct {
	From                string `json:"from,omitempty"`
	ID                  string `json:"id,omitempty"` // wamid of parent
	Forwarded           bool   `json:"forwarded,omitempty"`
	FrequentlyForwarded bool   `json:"frequently_forwarded,omitempty"`
}

type MetaStatus struct {
	ID           string            `json:"id"` // wamid.HBg...
	Status       string            `json:"status"` // sent, delivered, read, failed
	Timestamp    string            `json:"timestamp"`
	RecipientID  string            `json:"recipient_id"`
	Conversation *MetaStatusConv   `json:"conversation,omitempty"`
	Errors       []MetaStatusError `json:"errors,omitempty"`
}

type MetaStatusConv struct {
	ID     string `json:"id"`
	Origin struct {
		Type string `json:"type"`
	} `json:"origin"`
}

type MetaStatusError struct {
	Code      int    `json:"code"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	ErrorData struct {
		Details string `json:"details"`
	} `json:"error_data"`
}
