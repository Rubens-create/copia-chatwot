package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
)

type WebhookRepository interface {
	ListActiveWebhooks(ctx context.Context, accountID, inboxID int, event string) ([]model.Webhook, error)
	Create(ctx context.Context, wh *model.Webhook) (*model.Webhook, error)
	List(ctx context.Context, accountID int) ([]model.Webhook, error)
	LogDeliveryAttempt(ctx context.Context, attempt *model.WebhookDeliveryAttempt) (*model.WebhookDeliveryAttempt, error)
	UpdateDeliveryAttempt(ctx context.Context, id int64, attempts int, status, lastError string, responseCode int, nextAttemptAt *time.Time) error
}

type webhookRepository struct {
	db *database.DB
}

func NewWebhookRepository(db *database.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) ListActiveWebhooks(ctx context.Context, accountID, inboxID int, event string) ([]model.Webhook, error) {
	query := `
		SELECT id, account_id, inbox_id, url, webhook_type, subscriptions, name, secret, created_at, updated_at
		FROM webhooks
		WHERE (account_id IS NULL OR account_id = $1)
		  AND (inbox_id IS NULL OR inbox_id = $2)
	`
	rows, err := r.db.Pool.Query(ctx, query, accountID, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	webhooks := make([]model.Webhook, 0)
	for rows.Next() {
		var wh model.Webhook
		err := rows.Scan(&wh.ID, &wh.AccountID, &wh.InboxID, &wh.URL, &wh.WebhookType, &wh.Subscriptions, &wh.Name, &wh.Secret, &wh.CreatedAt, &wh.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		// Filter by subscription event if specified
		if len(wh.Subscriptions) > 0 && string(wh.Subscriptions) != "null" {
			var subs []string
			if err := json.Unmarshal(wh.Subscriptions, &subs); err == nil && len(subs) > 0 {
				subscribed := false
				for _, s := range subs {
					if s == event || s == "*" {
						subscribed = true
						break
					}
				}
				if !subscribed {
					continue
				}
			}
		}

		webhooks = append(webhooks, wh)
	}

	return webhooks, nil
}

func (r *webhookRepository) Create(ctx context.Context, wh *model.Webhook) (*model.Webhook, error) {
	now := time.Now().UTC()
	query := `
		INSERT INTO webhooks (account_id, inbox_id, url, webhook_type, subscriptions, name, secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id, account_id, inbox_id, url, webhook_type, subscriptions, name, secret, created_at, updated_at
	`
	subs := wh.Subscriptions
	if len(subs) == 0 {
		subs = []byte(`["message_created", "message_updated", "conversation_created", "conversation_updated"]`)
	}

	row := r.db.Pool.QueryRow(ctx, query, wh.AccountID, wh.InboxID, wh.URL, wh.WebhookType, subs, wh.Name, wh.Secret, now)

	var created model.Webhook
	err := row.Scan(&created.ID, &created.AccountID, &created.InboxID, &created.URL, &created.WebhookType, &created.Subscriptions, &created.Name, &created.Secret, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &created, nil
}

func (r *webhookRepository) List(ctx context.Context, accountID int) ([]model.Webhook, error) {
	query := `
		SELECT id, account_id, inbox_id, url, webhook_type, subscriptions, name, secret, created_at, updated_at
		FROM webhooks
		WHERE account_id IS NULL OR account_id = $1
		ORDER BY id ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	list := make([]model.Webhook, 0)
	for rows.Next() {
		var wh model.Webhook
		if err := rows.Scan(&wh.ID, &wh.AccountID, &wh.InboxID, &wh.URL, &wh.WebhookType, &wh.Subscriptions, &wh.Name, &wh.Secret, &wh.CreatedAt, &wh.UpdatedAt); err == nil {
			list = append(list, wh)
		}
	}
	return list, nil
}

func (r *webhookRepository) LogDeliveryAttempt(ctx context.Context, attempt *model.WebhookDeliveryAttempt) (*model.WebhookDeliveryAttempt, error) {
	now := time.Now().UTC()
	query := `
		INSERT INTO webhook_delivery_attempts (webhook_id, event, payload, url, attempts, status, last_error, response_code, last_attempt_at, next_attempt_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING id, webhook_id, event, payload, url, attempts, status, last_error, response_code, last_attempt_at, next_attempt_at, created_at, updated_at
	`
	row := r.db.Pool.QueryRow(
		ctx, query,
		attempt.WebhookID, attempt.Event, attempt.Payload, attempt.URL, attempt.Attempts, attempt.Status,
		attempt.LastError, attempt.ResponseCode, attempt.LastAttemptAt, attempt.NextAttemptAt, now,
	)

	var created model.WebhookDeliveryAttempt
	err := row.Scan(
		&created.ID, &created.WebhookID, &created.Event, &created.Payload, &created.URL, &created.Attempts,
		&created.Status, &created.LastError, &created.ResponseCode, &created.LastAttemptAt, &created.NextAttemptAt,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to log delivery attempt: %w", err)
	}
	return &created, nil
}

func (r *webhookRepository) UpdateDeliveryAttempt(ctx context.Context, id int64, attempts int, status, lastError string, responseCode int, nextAttemptAt *time.Time) error {
	now := time.Now().UTC()
	query := `
		UPDATE webhook_delivery_attempts
		SET attempts = $1, status = $2, last_error = $3, response_code = $4, next_attempt_at = $5, last_attempt_at = $6, updated_at = $6
		WHERE id = $7
	`
	_, err := r.db.Pool.Exec(ctx, query, attempts, status, lastError, responseCode, nextAttemptAt, now, id)
	return err
}
