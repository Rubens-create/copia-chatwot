package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
)

type dummyWhatsAppService struct{}

func (d *dummyWhatsAppService) ProcessWebhookPayload(ctx context.Context, rawBody []byte) error {
	return nil
}

func computeSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestMetaWebhook_ValidSignature(t *testing.T) {
	secret := "my_meta_app_secret_123"
	body := []byte(`{"object":"whatsapp_business_account","entry":[]}`)

	validSig := computeSignature(body, secret)
	if !ValidMetaSignature(body, validSig, secret) {
		t.Errorf("expected signature to be valid")
	}

	invalidSig := "sha256=abcdef1234567890"
	if ValidMetaSignature(body, invalidSig, secret) {
		t.Errorf("expected signature to be invalid")
	}

	missingPrefix := strings.TrimPrefix(validSig, "sha256=")
	if ValidMetaSignature(body, missingPrefix, secret) {
		t.Errorf("expected signature without prefix to be invalid")
	}

	// Secret empty should allow all
	if !ValidMetaSignature(body, "", "") {
		t.Errorf("expected empty secret to bypass validation")
	}
}

func TestMetaWebhook_ReceiveWebhook_HMAC(t *testing.T) {
	cfg := &config.Config{
		MetaAppSecret:   "secret_key_456",
		MetaVerifyToken: "token_123",
	}

	handler := NewMetaWebhookHandler(cfg, &dummyWhatsAppService{})
	body := `{"object":"whatsapp_business_account"}`

	// 1. Valid Signature -> 200 OK
	req := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", strings.NewReader(body))
	req.Header.Set("X-Hub-Signature-256", computeSignature([]byte(body), cfg.MetaAppSecret))
	w := httptest.NewRecorder()
	handler.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK with valid HMAC, got %d", w.Code)
	}

	// 2. Invalid Signature -> 401 Unauthorized
	reqInvalid := httptest.NewRequest(http.MethodPost, "/webhooks/whatsapp", strings.NewReader(body))
	reqInvalid.Header.Set("X-Hub-Signature-256", "sha256=invalid_hash")
	wInvalid := httptest.NewRecorder()
	handler.HandleWebhook(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized with invalid HMAC, got %d", wInvalid.Code)
	}
}
