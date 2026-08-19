package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SendMessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

type WhatsAppClient interface {
	SendText(ctx context.Context, phoneNumberID, to, text string) (*SendMessageResponse, error)
	SendImage(ctx context.Context, phoneNumberID, to, imageURL, caption string) (*SendMessageResponse, error)
	SendAudio(ctx context.Context, phoneNumberID, to, audioURL string) (*SendMessageResponse, error)
	SendVideo(ctx context.Context, phoneNumberID, to, videoURL, caption string) (*SendMessageResponse, error)
	SendDocument(ctx context.Context, phoneNumberID, to, docURL, filename, caption string) (*SendMessageResponse, error)
	SendLocation(ctx context.Context, phoneNumberID, to string, lat, long float64, name, address string) (*SendMessageResponse, error)
	SendTemplate(ctx context.Context, phoneNumberID, to, templateName, languageCode string, components []interface{}) (*SendMessageResponse, error)
}

type metaClient struct {
	accessToken string
	apiVersion  string
	httpClient  *http.Client
}

func NewClient(accessToken, apiVersion string) WhatsAppClient {
	if apiVersion == "" {
		apiVersion = "v19.0"
	}
	return &metaClient{
		accessToken: accessToken,
		apiVersion:  apiVersion,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *metaClient) sendRequest(ctx context.Context, phoneNumberID string, payload map[string]interface{}) (*SendMessageResponse, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, phoneNumberID)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute meta request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("meta api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode meta response: %w", err)
	}

	return &result, nil
}

func (c *metaClient) SendText(ctx context.Context, phoneNumberID, to, text string) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "text",
		"text": map[string]string{
			"body": text,
		},
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendImage(ctx context.Context, phoneNumberID, to, imageURL, caption string) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "image",
		"image": map[string]string{
			"link":    imageURL,
			"caption": caption,
		},
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendAudio(ctx context.Context, phoneNumberID, to, audioURL string) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "audio",
		"audio": map[string]string{
			"link": audioURL,
		},
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendVideo(ctx context.Context, phoneNumberID, to, videoURL, caption string) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "video",
		"video": map[string]string{
			"link":    videoURL,
			"caption": caption,
		},
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendDocument(ctx context.Context, phoneNumberID, to, docURL, filename, caption string) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "document",
		"document": map[string]string{
			"link":     docURL,
			"filename": filename,
			"caption":  caption,
		},
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendLocation(ctx context.Context, phoneNumberID, to string, lat, long float64, name, address string) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	locationData := map[string]interface{}{
		"latitude":  lat,
		"longitude": long,
	}
	if name != "" {
		locationData["name"] = name
	}
	if address != "" {
		locationData["address"] = address
	}
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "location",
		"location":          locationData,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendTemplate(ctx context.Context, phoneNumberID, to, templateName, languageCode string, components []interface{}) (*SendMessageResponse, error) {
	toDigits := strings.TrimPrefix(NormalizePhoneNumber(to), "+")
	if languageCode == "" {
		languageCode = "pt_BR"
	}
	templateData := map[string]interface{}{
		"name": templateName,
		"language": map[string]string{
			"code": languageCode,
		},
	}
	if len(components) > 0 {
		templateData["components"] = components
	}
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toDigits,
		"type":              "template",
		"template":          templateData,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}
