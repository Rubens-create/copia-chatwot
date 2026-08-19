package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/service"
)

type dummyConversationService struct {
	lastParams service.SendMessageParams
}

func (d *dummyConversationService) ListConversations(ctx context.Context, accountID, inboxID int, status *int, limit, offset int) ([]model.Conversation, error) {
	return []model.Conversation{{ID: 1, AccountID: 1}}, nil
}

func (d *dummyConversationService) GetConversation(ctx context.Context, id int) (*model.Conversation, error) {
	return &model.Conversation{ID: id, AccountID: 1}, nil
}

func (d *dummyConversationService) GetMessages(ctx context.Context, conversationID int, limit, offset int) ([]model.Message, error) {
	content := "Hello"
	return []model.Message{{ID: 1, ConversationID: conversationID, Content: &content}}, nil
}

func (d *dummyConversationService) SendMessage(ctx context.Context, conversationID int, params service.SendMessageParams) (*model.Message, error) {
	d.lastParams = params
	return &model.Message{ID: 100, ConversationID: conversationID, Content: &params.Content}, nil
}

func TestHandleMessagesMultipartPassesVoiceFlagToAudioAttachment(t *testing.T) {
	convService := &dummyConversationService{}
	convHandler := NewConversationHandler(convService)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("is_voice", "true"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("attachments[]", "gravacao.ogg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("OggS-test")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/1/conversations/42/messages", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	convHandler.HandleMessages(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}
	if len(convService.lastParams.Attachments) != 1 {
		t.Fatalf("expected one attachment, got %d", len(convService.lastParams.Attachments))
	}
	attachment := convService.lastParams.Attachments[0]
	if attachment.FileType != 1 || !attachment.IsVoice {
		t.Fatalf("voice attachment not propagated: %#v", attachment)
	}
}

func TestAuthMiddleware(t *testing.T) {
	token := "secret_api_token_xyz"
	auth := AuthMiddleware(token)

	handler := auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	// 1. Missing Token -> 401
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/conversations", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing token, got %d", w.Code)
	}

	// 2. Invalid Token -> 401
	reqInvalid := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/conversations", nil)
	reqInvalid.Header.Set("api_access_token", "wrong_token")
	wInvalid := httptest.NewRecorder()
	handler.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for wrong token, got %d", wInvalid.Code)
	}

	// 3. Valid api_access_token header -> 200
	reqValidHeader := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/conversations", nil)
	reqValidHeader.Header.Set("api_access_token", token)
	wValidHeader := httptest.NewRecorder()
	handler.ServeHTTP(wValidHeader, reqValidHeader)
	if wValidHeader.Code != http.StatusOK {
		t.Errorf("expected 200 OK for valid header, got %d", wValidHeader.Code)
	}

	// 4. Valid Authorization Bearer header -> 200
	reqValidBearer := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/conversations", nil)
	reqValidBearer.Header.Set("Authorization", "Bearer "+token)
	wValidBearer := httptest.NewRecorder()
	handler.ServeHTTP(wValidBearer, reqValidBearer)
	if wValidBearer.Code != http.StatusOK {
		t.Errorf("expected 200 OK for Bearer token, got %d", wValidBearer.Code)
	}

	// 5. Valid Query Param -> 200
	reqValidQuery := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/conversations?api_access_token="+token, nil)
	wValidQuery := httptest.NewRecorder()
	handler.ServeHTTP(wValidQuery, reqValidQuery)
	if wValidQuery.Code != http.StatusOK {
		t.Errorf("expected 200 OK for query token, got %d", wValidQuery.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	cors := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Preflight OPTIONS request
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/accounts/1/conversations/42/messages", nil)
	w := httptest.NewRecorder()
	cors.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK for OPTIONS preflight, got %d", w.Code)
	}

	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	if !bytes.Contains([]byte(allowedHeaders), []byte("api_access_token")) {
		t.Errorf("expected api_access_token in Access-Control-Allow-Headers, got %s", allowedHeaders)
	}
}

func TestChatwootV1_ConversationRoutes(t *testing.T) {
	convHandler := NewConversationHandler(&dummyConversationService{})

	// 1. List Conversations v1: GET /api/v1/accounts/1/conversations
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/1/conversations", nil)
	wList := httptest.NewRecorder()
	convHandler.ListConversations(wList, reqList)
	if wList.Code != http.StatusOK {
		t.Errorf("expected 200 OK for list conversations, got %d", wList.Code)
	}

	// 2. Post Message with Attachment: POST /api/v1/accounts/1/conversations/42/messages
	payload := service.SendMessageParams{
		Content: "Foto do produto",
		Attachments: []service.AttachmentParam{
			{
				FileType:      0,
				ExternalURL:   "https://example.com/foto.jpg",
				FallbackTitle: "foto.jpg",
			},
		},
	}
	body, _ := json.Marshal(payload)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/1/conversations/42/messages", bytes.NewReader(body))
	wPost := httptest.NewRecorder()
	convHandler.HandleMessages(wPost, reqPost)
	if wPost.Code != http.StatusCreated {
		t.Errorf("expected 201 Created for post message with attachment, got %d", wPost.Code)
	}

	// 3. Post Message with Template: POST /api/v1/accounts/1/conversations/42/messages
	templatePayload := service.SendMessageParams{
		Template: &service.TemplateParam{
			Name:     "boas_vindas",
			Language: "pt_BR",
		},
	}
	templateBody, _ := json.Marshal(templatePayload)
	reqTemplate := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/1/conversations/42/messages", bytes.NewReader(templateBody))
	wTemplate := httptest.NewRecorder()
	convHandler.HandleMessages(wTemplate, reqTemplate)
	if wTemplate.Code != http.StatusCreated {
		t.Errorf("expected 201 Created for post template message, got %d", wTemplate.Code)
	}
}
