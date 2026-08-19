package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
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
	SendImage(ctx context.Context, phoneNumberID, to, mediaIDOrURL, caption string) (*SendMessageResponse, error)
	SendAudio(ctx context.Context, phoneNumberID, to, mediaIDOrURL string) (*SendMessageResponse, error)
	SendVideo(ctx context.Context, phoneNumberID, to, mediaIDOrURL, caption string) (*SendMessageResponse, error)
	SendDocument(ctx context.Context, phoneNumberID, to, mediaIDOrURL, filename, caption string) (*SendMessageResponse, error)
	SendLocation(ctx context.Context, phoneNumberID, to string, lat, long float64, name, address string) (*SendMessageResponse, error)
	SendTemplate(ctx context.Context, phoneNumberID, to, templateName, languageCode string, components []interface{}) (*SendMessageResponse, error)
	UploadMedia(ctx context.Context, phoneNumberID, filename, mimeType string, data []byte) (string, error)
	UpdateCredentials(accessToken, apiVersion string)
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
			Timeout: 25 * time.Second,
		},
	}
}

func (c *metaClient) UpdateCredentials(accessToken, apiVersion string) {
	if accessToken != "" {
		c.accessToken = accessToken
	}
	if apiVersion != "" {
		c.apiVersion = apiVersion
	}
}

func (c *metaClient) UploadMedia(ctx context.Context, phoneNumberID, filename, mimeType string, data []byte) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/media", c.apiVersion, phoneNumberID)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if err := w.WriteField("messaging_product", "whatsapp"); err != nil {
		return "", fmt.Errorf("failed to write messaging_product field: %w", err)
	}

	if mimeType != "" {
		if err := w.WriteField("type", mimeType); err != nil {
			return "", fmt.Errorf("failed to write type field: %w", err)
		}
	}

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	if mimeType != "" {
		partHeader.Set("Content-Type", mimeType)
	}
	part, err := w.CreatePart(partHeader)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(data); err != nil {
		return "", fmt.Errorf("failed to write file data: %w", err)
	}

	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &b)
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("meta media upload error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var uploadResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &uploadResult); err != nil {
		return "", fmt.Errorf("failed to decode upload response: %w", err)
	}

	if uploadResult.ID == "" {
		return "", fmt.Errorf("meta returned empty media id: %s", string(respBody))
	}

	return uploadResult.ID, nil
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
	toFormatted := NormalizePhoneNumber(to)
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toFormatted,
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": false,
			"body":        text,
		},
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendImage(ctx context.Context, phoneNumberID, to, mediaIDOrURL, caption string) (*SendMessageResponse, error) {
	toFormatted := NormalizePhoneNumber(to)
	imgObj := map[string]interface{}{}
	if strings.HasPrefix(mediaIDOrURL, "http://") || strings.HasPrefix(mediaIDOrURL, "https://") {
		imgObj["link"] = mediaIDOrURL
	} else {
		imgObj["id"] = mediaIDOrURL
	}
	if caption != "" {
		imgObj["caption"] = caption
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toFormatted,
		"type":              "image",
		"image":             imgObj,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendAudio(ctx context.Context, phoneNumberID, to, mediaIDOrURL string) (*SendMessageResponse, error) {
	toFormatted := NormalizePhoneNumber(to)
	audioObj := map[string]interface{}{}
	if strings.HasPrefix(mediaIDOrURL, "http://") || strings.HasPrefix(mediaIDOrURL, "https://") {
		audioObj["link"] = mediaIDOrURL
	} else {
		audioObj["id"] = mediaIDOrURL
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toFormatted,
		"type":              "audio",
		"audio":             audioObj,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendVideo(ctx context.Context, phoneNumberID, to, mediaIDOrURL, caption string) (*SendMessageResponse, error) {
	toFormatted := NormalizePhoneNumber(to)
	vidObj := map[string]interface{}{}
	if strings.HasPrefix(mediaIDOrURL, "http://") || strings.HasPrefix(mediaIDOrURL, "https://") {
		vidObj["link"] = mediaIDOrURL
	} else {
		vidObj["id"] = mediaIDOrURL
	}
	if caption != "" {
		vidObj["caption"] = caption
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toFormatted,
		"type":              "video",
		"video":             vidObj,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendDocument(ctx context.Context, phoneNumberID, to, mediaIDOrURL, filename, caption string) (*SendMessageResponse, error) {
	toFormatted := NormalizePhoneNumber(to)
	docObj := map[string]interface{}{}
	if strings.HasPrefix(mediaIDOrURL, "http://") || strings.HasPrefix(mediaIDOrURL, "https://") {
		docObj["link"] = mediaIDOrURL
	} else {
		docObj["id"] = mediaIDOrURL
	}
	if filename != "" {
		docObj["filename"] = filename
	}
	if caption != "" {
		docObj["caption"] = caption
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toFormatted,
		"type":              "document",
		"document":          docObj,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendLocation(ctx context.Context, phoneNumberID, to string, lat, long float64, name, address string) (*SendMessageResponse, error) {
	toFormatted := NormalizePhoneNumber(to)
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
		"to":                toFormatted,
		"type":              "location",
		"location":          locationData,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}

func (c *metaClient) SendTemplate(ctx context.Context, phoneNumberID, to, templateName, languageCode string, components []interface{}) (*SendMessageResponse, error) {
	toFormatted := NormalizePhoneNumber(to)
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
		"to":                toFormatted,
		"type":              "template",
		"template":          templateData,
	}
	return c.sendRequest(ctx, phoneNumberID, payload)
}
