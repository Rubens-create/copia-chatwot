package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/service"
)

type ConversationHandler struct {
	convService service.ConversationService
}

func NewConversationHandler(convService service.ConversationService) *ConversationHandler {
	return &ConversationHandler{convService: convService}
}

func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	accountID, _ := strconv.Atoi(q.Get("account_id"))
	if accountID == 0 {
		accountID, _, _ = parseConversationRoute(r.URL.Path)
	}
	inboxID, _ := strconv.Atoi(q.Get("inbox_id"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var statusPtr *int
	if statusStr := q.Get("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}

	conversations, err := h.convService.ListConversations(r.Context(), accountID, inboxID, statusPtr, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"failed to list conversations"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": conversations,
	})
}

func (h *ConversationHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	_, convID, _ := parseConversationRoute(r.URL.Path)
	if convID <= 0 {
		convID = extractIDFromPath(r.URL.Path, "/api/conversations/")
	}
	if convID <= 0 {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	conv, err := h.convService.GetConversation(r.Context(), convID)
	if err != nil {
		http.Error(w, `{"error":"error retrieving conversation"}`, http.StatusInternalServerError)
		return
	}
	if conv == nil {
		http.Error(w, `{"error":"conversation not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"data": conv,
	})
}

func (h *ConversationHandler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	_, convID, _ := parseConversationRoute(r.URL.Path)
	if convID <= 0 {
		// Legacy path: /api/conversations/{id}/messages
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 3 {
			convID, _ = strconv.Atoi(parts[2])
		}
	}

	if convID <= 0 {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

		messages, err := h.convService.GetMessages(r.Context(), convID, limit, offset)
		if err != nil {
			http.Error(w, `{"error":"failed to list messages"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": messages,
		})

	case http.MethodPost:
		var req service.SendMessageParams
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request payload: `+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		// Validation: at least content, attachment or template must be present
		hasContent := strings.TrimSpace(req.Content) != ""
		hasAttachments := len(req.Attachments) > 0
		hasTemplate := (req.Template != nil && req.Template.Name != "") || req.TemplateName != ""

		if !hasContent && !hasAttachments && !hasTemplate {
			http.Error(w, `{"error":"content, attachments or template is required"}`, http.StatusBadRequest)
			return
		}

		msg, err := h.convService.SendMessage(r.Context(), convID, req)
		if err != nil {
			http.Error(w, `{"error":"failed to send message: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": msg,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// parseConversationRoute extracts account_id, conversation_id and resource from Chatwoot v1 paths.
// Example: /api/v1/accounts/1/conversations/42/messages -> accountID=1, convID=42, resource="messages"
func parseConversationRoute(path string) (accountID int, convID int, resource string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// Case 1: api/v1/accounts/{account_id}/conversations[/{conv_id}[/{resource}]]
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "accounts" {
		accountID, _ = strconv.Atoi(parts[3])
		if len(parts) >= 6 && parts[4] == "conversations" {
			convID, _ = strconv.Atoi(parts[5])
			if len(parts) >= 7 {
				resource = parts[6]
			}
		}
		return
	}

	// Case 2: api/conversations/{conv_id}[/{resource}]
	if len(parts) >= 2 && parts[0] == "api" && parts[1] == "conversations" {
		if len(parts) >= 3 {
			convID, _ = strconv.Atoi(parts[2])
			if len(parts) >= 4 {
				resource = parts[3]
			}
		}
		return
	}

	return 0, 0, ""
}

func extractIDFromPath(path, prefix string) int {
	clean := strings.TrimPrefix(path, prefix)
	clean = strings.Split(clean, "/")[0]
	id, err := strconv.Atoi(clean)
	if err != nil {
		return 0
	}
	return id
}
