package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/queue"
)

type HealthHandler struct {
	db    *database.DB
	queue *queue.RedisQueue
}

func NewHealthHandler(db *database.DB, q *queue.RedisQueue) *HealthHandler {
	return &HealthHandler{db: db, queue: q}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"system": "chatwoot-lite-whatsapp-gateway",
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allOk := true

	if err := h.db.Ping(ctx); err != nil {
		checks["postgres"] = "error: " + err.Error()
		allOk = false
	} else {
		checks["postgres"] = "connected"
	}

	if err := h.queue.Ping(ctx); err != nil {
		checks["redis"] = "error: " + err.Error()
		allOk = false
	} else {
		checks["redis"] = "connected"
	}

	w.Header().Set("Content-Type", "application/json")
	if !allOk {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": func() string {
			if allOk {
				return "ready"
			}
			return "unready"
		}(),
		"checks": checks,
	})
}
