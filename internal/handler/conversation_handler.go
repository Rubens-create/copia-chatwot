package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
	accountID, _, _ := parseConversationRoute(r.URL.Path)
	if accountID == 0 {
		accountID, _ = strconv.Atoi(r.URL.Query().Get("account_id"))
	}
	if accountID == 0 {
		accountID = 1
	}

	inboxID, _ := strconv.Atoi(r.URL.Query().Get("inbox_id"))
	var statusPtr *int
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusPtr = &s
		}
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

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
	_, convID, _ := parseConversationRoute(r.URL.Path)
	if convID <= 0 {
		// Legacy path: /api/conversations/{id}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 3 {
			convID, _ = strconv.Atoi(parts[2])
		}
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

		ct := r.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				http.Error(w, `{"error":"failed to parse multipart form: `+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			req.Content = r.FormValue("content")
			if r.FormValue("private") == "true" {
				req.Private = true
			}

			var fileHeaders []*multipart.FileHeader
			if r.MultipartForm != nil && r.MultipartForm.File != nil {
				for _, key := range []string{"attachments[]", "attachments", "file", "attachment", "files[]"} {
					if fhs, ok := r.MultipartForm.File[key]; ok {
						fileHeaders = append(fileHeaders, fhs...)
					}
				}
			}

			for _, fh := range fileHeaders {
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}

				mimeType := fh.Header.Get("Content-Type")
				if mimeType == "" {
					mimeType = http.DetectContentType(data)
				}

				fileType := 3 // document
				if strings.HasPrefix(mimeType, "image/") {
					fileType = 0
				} else if strings.HasPrefix(mimeType, "audio/") {
					fileType = 1
				} else if strings.HasPrefix(mimeType, "video/") {
					fileType = 2
				}

				dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
				req.Attachments = append(req.Attachments, service.AttachmentParam{
					FileType:      fileType,
					DataURL:       dataURL,
					FallbackTitle: fh.Filename,
				})
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request payload: `+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
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
