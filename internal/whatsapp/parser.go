package whatsapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
)

type ParsedIncomingMessage struct {
	MetaMessageID  string
	SenderPhone    string
	SenderName     string
	Timestamp      time.Time
	Type           string
	Content        string
	ContentType    int
	Attachments    []ParsedAttachment
	ContextReplyID string
	RawPayload     json.RawMessage
	PhoneNumberID  string
	DisplayPhone   string
}

type ParsedAttachment struct {
	FileType        int
	ExternalURL     string
	CoordinatesLat  float64
	CoordinatesLong float64
	FallbackTitle   string
	Extension       string
	Meta            map[string]interface{}
}

type ParsedStatusUpdate struct {
	MetaMessageID  string
	Status         int // model.MessageStatus...
	StatusStr      string
	Timestamp      time.Time
	RecipientPhone string
	ErrorDetails   string
	PhoneNumberID  string
	DisplayPhone   string
}

type ParsedWebhookResult struct {
	Messages []ParsedIncomingMessage
	Statuses []ParsedStatusUpdate
}

func ParseWebhookPayload(body []byte) (*ParsedWebhookResult, error) {
	var payload model.MetaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json payload: %w", err)
	}

	result := &ParsedWebhookResult{
		Messages: make([]ParsedIncomingMessage, 0),
		Statuses: make([]ParsedStatusUpdate, 0),
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}

			val := change.Value
			phoneID := val.Metadata.PhoneNumberID
			displayPhone := NormalizePhoneNumber(val.Metadata.DisplayPhoneNumber)

			// Map sender names by wa_id
			contactNames := make(map[string]string)
			for _, contact := range val.Contacts {
				contactNames[contact.WaID] = contact.Profile.Name
			}

			// Process incoming messages
			for _, msg := range val.Messages {
				parsedMsg := parseSingleMessage(msg, phoneID, displayPhone, contactNames, body)
				result.Messages = append(result.Messages, parsedMsg)
			}

			// Process status updates
			for _, status := range val.Statuses {
				parsedStatus := parseSingleStatus(status, phoneID, displayPhone)
				result.Statuses = append(result.Statuses, parsedStatus)
			}
		}
	}

	return result, nil
}

func parseSingleMessage(msg model.MetaMessage, phoneID, displayPhone string, contactNames map[string]string, rawBody []byte) ParsedIncomingMessage {
	senderPhone := NormalizePhoneNumber(msg.From)
	senderName := contactNames[msg.From]
	if senderName == "" {
		senderName = senderPhone
	}

	t := time.Now().UTC()
	if tsInt, err := strconv.ParseInt(msg.Timestamp, 10, 64); err == nil && tsInt > 0 {
		t = time.Unix(tsInt, 0).UTC()
	}

	parsed := ParsedIncomingMessage{
		MetaMessageID: msg.ID,
		SenderPhone:   senderPhone,
		SenderName:    senderName,
		Timestamp:     t,
		Type:          msg.Type,
		ContentType:   model.ContentTypeText,
		PhoneNumberID: phoneID,
		DisplayPhone:  displayPhone,
		RawPayload:    rawBody,
		Attachments:   make([]ParsedAttachment, 0),
	}

	if msg.Context != nil {
		parsed.ContextReplyID = msg.Context.ID
	}

	switch msg.Type {
	case "text":
		if msg.Text != nil {
			parsed.Content = msg.Text.Body
		}

	case "image":
		if msg.Image != nil {
			parsed.Content = msg.Image.Caption
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:      model.AttachmentTypeImage,
				ExternalURL:   msg.Image.ID, // Meta Media ID for retrieval
				FallbackTitle: "Image",
				Extension:     mimeToExt(msg.Image.MimeType, "jpg"),
				Meta: map[string]interface{}{
					"mime_type": msg.Image.MimeType,
					"sha256":    msg.Image.Sha256,
					"media_id":  msg.Image.ID,
				},
			})
		}

	case "audio":
		if msg.Audio != nil {
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:      model.AttachmentTypeAudio,
				ExternalURL:   msg.Audio.ID,
				FallbackTitle: "Audio",
				Extension:     mimeToExt(msg.Audio.MimeType, "ogg"),
				Meta: map[string]interface{}{
					"mime_type": msg.Audio.MimeType,
					"voice":     msg.Audio.Voice,
					"media_id":  msg.Audio.ID,
				},
			})
		}

	case "video":
		if msg.Video != nil {
			parsed.Content = msg.Video.Caption
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:      model.AttachmentTypeVideo,
				ExternalURL:   msg.Video.ID,
				FallbackTitle: "Video",
				Extension:     mimeToExt(msg.Video.MimeType, "mp4"),
				Meta: map[string]interface{}{
					"mime_type": msg.Video.MimeType,
					"sha256":    msg.Video.Sha256,
					"media_id":  msg.Video.ID,
				},
			})
		}

	case "document":
		if msg.Document != nil {
			parsed.Content = msg.Document.Caption
			filename := msg.Document.Filename
			if filename == "" {
				filename = "Document"
			}
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:      model.AttachmentTypeFile,
				ExternalURL:   msg.Document.ID,
				FallbackTitle: filename,
				Extension:     mimeToExt(msg.Document.MimeType, "pdf"),
				Meta: map[string]interface{}{
					"mime_type": msg.Document.MimeType,
					"sha256":    msg.Document.Sha256,
					"filename":  filename,
					"media_id":  msg.Document.ID,
				},
			})
		}

	case "location":
		if msg.Location != nil {
			title := msg.Location.Name
			if title == "" {
				title = msg.Location.Address
			}
			if title == "" {
				title = fmt.Sprintf("Location (%.4f, %.4f)", msg.Location.Latitude, msg.Location.Longitude)
			}
			parsed.Content = title
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:        model.AttachmentTypeLocation,
				CoordinatesLat:  msg.Location.Latitude,
				CoordinatesLong: msg.Location.Longitude,
				FallbackTitle:   title,
				Meta: map[string]interface{}{
					"name":      msg.Location.Name,
					"address":   msg.Location.Address,
					"latitude":  msg.Location.Latitude,
					"longitude": msg.Location.Longitude,
				},
			})
		}

	case "contacts":
		if len(msg.Contacts) > 0 {
			c := msg.Contacts[0]
			title := c.Name.FormattedName
			phone := ""
			if len(c.Phones) > 0 {
				phone = c.Phones[0].Phone
			}
			parsed.Content = fmt.Sprintf("Contact: %s (%s)", title, phone)
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:      model.AttachmentTypeContact,
				FallbackTitle: title,
				Meta: map[string]interface{}{
					"name":   c.Name,
					"phones": c.Phones,
					"emails": c.Emails,
				},
			})
		}

	case "interactive":
		if msg.Interactive != nil {
			if msg.Interactive.ButtonReply != nil {
				parsed.Content = msg.Interactive.ButtonReply.Title
			} else if msg.Interactive.ListReply != nil {
				parsed.Content = msg.Interactive.ListReply.Title
			}
		}

	case "sticker":
		if msg.Sticker != nil {
			parsed.Attachments = append(parsed.Attachments, ParsedAttachment{
				FileType:      model.AttachmentTypeImage,
				ExternalURL:   msg.Sticker.ID,
				FallbackTitle: "Sticker",
				Extension:     mimeToExt(msg.Sticker.MimeType, "webp"),
				Meta: map[string]interface{}{
					"mime_type": msg.Sticker.MimeType,
					"sha256":    msg.Sticker.Sha256,
					"media_id":  msg.Sticker.ID,
				},
			})
		}
	default:
		parsed.Content = fmt.Sprintf("[%s message]", msg.Type)
	}

	return parsed
}

func parseSingleStatus(status model.MetaStatus, phoneID, displayPhone string) ParsedStatusUpdate {
	t := time.Now().UTC()
	if tsInt, err := strconv.ParseInt(status.Timestamp, 10, 64); err == nil && tsInt > 0 {
		t = time.Unix(tsInt, 0).UTC()
	}

	statusEnum := model.MessageStatusSent
	switch status.Status {
	case "delivered":
		statusEnum = model.MessageStatusDelivered
	case "read":
		statusEnum = model.MessageStatusRead
	case "failed":
		statusEnum = model.MessageStatusFailed
	case "sent":
		statusEnum = model.MessageStatusSent
	}

	errDetails := ""
	if len(status.Errors) > 0 {
		errDetails = fmt.Sprintf("Code %d: %s - %s", status.Errors[0].Code, status.Errors[0].Title, status.Errors[0].Message)
	}

	return ParsedStatusUpdate{
		MetaMessageID:  status.ID,
		Status:         statusEnum,
		StatusStr:      status.Status,
		Timestamp:      t,
		RecipientPhone: NormalizePhoneNumber(status.RecipientID),
		ErrorDetails:   errDetails,
		PhoneNumberID:  phoneID,
		DisplayPhone:   displayPhone,
	}
}

func mimeToExt(mime, defaultExt string) string {
	switch mime {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "audio/ogg", "audio/ogg; codecs=opus":
		return "ogg"
	case "audio/mp4", "audio/m4a":
		return "m4a"
	case "audio/mpeg", "audio/mp3":
		return "mp3"
	case "video/mp4":
		return "mp4"
	case "application/pdf":
		return "pdf"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	default:
		return defaultExt
	}
}
