package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/service"
)

type MetaWebhookHandler struct {
	cfg       *config.Config
	waService service.WhatsAppService
}

func NewMetaWebhookHandler(cfg *config.Config, waService service.WhatsAppService) *MetaWebhookHandler {
	return &MetaWebhookHandler{
		cfg:       cfg,
		waService: waService,
	}
}

// HandleWebhook routes GET (verification) and POST (incoming payload)
func (h *MetaWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.verifyWebhook(w, r)
	case http.MethodPost:
		h.receiveWebhook(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *MetaWebhookHandler) verifyWebhook(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	mode := query.Get("hub.mode")
	token := query.Get("hub.verify_token")
	challenge := query.Get("hub.challenge")

	if mode == "subscribe" && token == h.cfg.MetaVerifyToken {
		log.Printf("[MetaWebhook] Webhook verified successfully with challenge token")
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(challenge))
		return
	}

	log.Printf("[MetaWebhook] Verification failed. Token mismatch or invalid mode: mode=%s", mode)
	http.Error(w, "Forbidden: Invalid verify token", http.StatusForbidden)
}

func (h *MetaWebhookHandler) receiveWebhook(w http.ResponseWriter, r *http.Request) {
	// Limit body to 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[MetaWebhook] Failed to read webhook body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 1. HMAC-SHA256 Signature Validation (X-Hub-Signature-256)
	signature := r.Header.Get("X-Hub-Signature-256")
	if !ValidMetaSignature(body, signature, h.cfg.MetaAppSecret) {
		log.Printf("[MetaWebhook] Unauthorized: Invalid X-Hub-Signature-256 header")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid signature",
		})
		return
	}

	// 2. Process payload
	if err := h.waService.ProcessWebhookPayload(r.Context(), body); err != nil {
		log.Printf("[MetaWebhook] Error processing payload: %v", err)
	}

	// Always respond HTTP 200 quickly to Meta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "received",
	})
}

// ValidMetaSignature validates Meta's X-Hub-Signature-256 using constant-time comparison
func ValidMetaSignature(body []byte, signature, secret string) bool {
	if secret == "" {
		return true
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return subtle.ConstantTimeCompare(got, mac.Sum(nil)) == 1
}
