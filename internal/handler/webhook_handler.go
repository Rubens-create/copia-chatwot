package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/config"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/repository"
)

type WebhookHandler struct {
	cfg         *config.Config
	webhookRepo repository.WebhookRepository
}

func NewWebhookHandler(cfg *config.Config, webhookRepo repository.WebhookRepository) *WebhookHandler {
	return &WebhookHandler{
		cfg:         cfg,
		webhookRepo: webhookRepo,
	}
}

func (h *WebhookHandler) HandleWebhooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		webhooks, err := h.webhookRepo.List(r.Context(), h.cfg.DefaultAccountID)
		if err != nil {
			http.Error(w, `{"error":"failed to list webhooks"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": webhooks,
		})

	case http.MethodPost:
		var req struct {
			URL           string   `json:"url"`
			Name          string   `json:"name"`
			Secret        string   `json:"secret"`
			Subscriptions []string `json:"subscriptions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
			http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
			return
		}

		subsBytes, _ := json.Marshal(req.Subscriptions)
		if len(req.Subscriptions) == 0 {
			subsBytes = []byte(`["message_created", "message_updated", "conversation_created", "conversation_updated"]`)
		}

		wh := &model.Webhook{
			AccountID:     &h.cfg.DefaultAccountID,
			InboxID:       &h.cfg.DefaultInboxID,
			URL:           req.URL,
			WebhookType:   0,
			Subscriptions: subsBytes,
		}
		if req.Name != "" {
			wh.Name = &req.Name
		}
		if req.Secret != "" {
			wh.Secret = &req.Secret
		}

		created, err := h.webhookRepo.Create(r.Context(), wh)
		if err != nil {
			http.Error(w, `{"error":"failed to register webhook: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": created,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
