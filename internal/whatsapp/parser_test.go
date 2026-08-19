package whatsapp

import (
	"testing"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
)

func TestParseWebhookPayload_TextMessage(t *testing.T) {
	jsonPayload := `{
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
									"profile": { "name": "Maria Silva" },
									"wa_id": "5562999999999"
								}
							],
							"messages": [
								{
									"from": "5562999999999",
									"id": "wamid.HBgLMTEyMjMzNDQ1NQ==",
									"timestamp": "1723456789",
									"type": "text",
									"text": { "body": "Olá, preciso de suporte!" }
								}
							]
						},
						"field": "messages"
					}
				]
			}
		]
	}`

	result, err := ParseWebhookPayload([]byte(jsonPayload))
	if err != nil {
		t.Fatalf("ParseWebhookPayload failed: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}

	msg := result.Messages[0]
	if msg.MetaMessageID != "wamid.HBgLMTEyMjMzNDQ1NQ==" {
		t.Errorf("expected MetaMessageID wamid.HBgLMTEyMjMzNDQ1NQ==, got %s", msg.MetaMessageID)
	}
	if msg.SenderPhone != "+5562999999999" {
		t.Errorf("expected SenderPhone +5562999999999, got %s", msg.SenderPhone)
	}
	if msg.SenderName != "Maria Silva" {
		t.Errorf("expected SenderName Maria Silva, got %s", msg.SenderName)
	}
	if msg.Content != "Olá, preciso de suporte!" {
		t.Errorf("expected Content 'Olá, preciso de suporte!', got %s", msg.Content)
	}
	if msg.PhoneNumberID != "phone_123" {
		t.Errorf("expected PhoneNumberID phone_123, got %s", msg.PhoneNumberID)
	}
}

func TestParseWebhookPayload_MediaAndLocation(t *testing.T) {
	jsonPayload := `{
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
							"messages": [
								{
									"from": "5562999999999",
									"id": "wamid.IMAGE_123",
									"timestamp": "1723456789",
									"type": "image",
									"image": {
										"id": "media_img_999",
										"mime_type": "image/jpeg",
										"sha256": "abcdef",
										"caption": "Comprovante"
									}
								},
								{
									"from": "5562999999999",
									"id": "wamid.LOC_123",
									"timestamp": "1723456790",
									"type": "location",
									"location": {
										"latitude": -16.686891,
										"longitude": -49.264794,
										"name": "Escritório Central",
										"address": "Av. Principal, 100"
									}
								}
							]
						},
						"field": "messages"
					}
				]
			}
		]
	}`

	result, err := ParseWebhookPayload([]byte(jsonPayload))
	if err != nil {
		t.Fatalf("ParseWebhookPayload failed: %v", err)
	}

	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Messages))
	}

	// 1. Image
	imgMsg := result.Messages[0]
	if imgMsg.Content != "Comprovante" {
		t.Errorf("expected image caption 'Comprovante', got %s", imgMsg.Content)
	}
	if len(imgMsg.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(imgMsg.Attachments))
	}
	if imgMsg.Attachments[0].FileType != model.AttachmentTypeImage {
		t.Errorf("expected AttachmentTypeImage, got %d", imgMsg.Attachments[0].FileType)
	}

	// 2. Location
	locMsg := result.Messages[1]
	if locMsg.Content != "Escritório Central" {
		t.Errorf("expected location name 'Escritório Central', got %s", locMsg.Content)
	}
	if len(locMsg.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(locMsg.Attachments))
	}
	if locMsg.Attachments[0].CoordinatesLat != -16.686891 {
		t.Errorf("expected lat -16.686891, got %f", locMsg.Attachments[0].CoordinatesLat)
	}
}

func TestParseWebhookPayload_Statuses(t *testing.T) {
	jsonPayload := `{
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
							"statuses": [
								{
									"id": "wamid.SENT_MSG_1",
									"status": "delivered",
									"timestamp": "1723456800",
									"recipient_id": "5562999999999"
								},
								{
									"id": "wamid.SENT_MSG_2",
									"status": "read",
									"timestamp": "1723456810",
									"recipient_id": "5562999999999"
								}
							]
						},
						"field": "messages"
					}
				]
			}
		]
	}`

	result, err := ParseWebhookPayload([]byte(jsonPayload))
	if err != nil {
		t.Fatalf("ParseWebhookPayload failed: %v", err)
	}

	if len(result.Statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(result.Statuses))
	}

	if result.Statuses[0].Status != model.MessageStatusDelivered {
		t.Errorf("expected MessageStatusDelivered, got %d", result.Statuses[0].Status)
	}
	if result.Statuses[1].Status != model.MessageStatusRead {
		t.Errorf("expected MessageStatusRead, got %d", result.Statuses[1].Status)
	}
}
