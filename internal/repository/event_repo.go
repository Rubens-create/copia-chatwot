package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/jackc/pgx/v5"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *model.WhatsAppGatewayEvent) (*model.WhatsAppGatewayEvent, error)
	CreateEventTx(ctx context.Context, tx pgx.Tx, event *model.WhatsAppGatewayEvent) (*model.WhatsAppGatewayEvent, error)
	UpdateProcessed(ctx context.Context, eventID int64) error
}

type eventRepository struct {
	db *database.DB
}

func NewEventRepository(db *database.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) CreateEvent(ctx context.Context, event *model.WhatsAppGatewayEvent) (*model.WhatsAppGatewayEvent, error) {
	now := time.Now().UTC()
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = now
	}
	event.CreatedAt = now

	insertQuery := `
		INSERT INTO whatsapp_gateway_events (
			account_id, inbox_id, message_id, event_type, external_event_id,
			phone_number_id, business_account_id, raw_payload, received_at, processed_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (inbox_id, event_type, external_event_id)
		DO UPDATE SET processed_at = EXCLUDED.processed_at
		RETURNING id, account_id, inbox_id, message_id, event_type, external_event_id,
		          phone_number_id, business_account_id, raw_payload, received_at, processed_at, created_at
	`
	row := r.db.Pool.QueryRow(
		ctx, insertQuery,
		event.AccountID, event.InboxID, event.MessageID, event.EventType, event.ExternalEventID,
		event.PhoneNumberID, event.BusinessAccountID, event.RawPayload, event.ReceivedAt, event.ProcessedAt, event.CreatedAt,
	)

	var res model.WhatsAppGatewayEvent
	err := row.Scan(
		&res.ID, &res.AccountID, &res.InboxID, &res.MessageID, &res.EventType, &res.ExternalEventID,
		&res.PhoneNumberID, &res.BusinessAccountID, &res.RawPayload, &res.ReceivedAt, &res.ProcessedAt, &res.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert gateway event: %w", err)
	}
	return &res, nil
}

func (r *eventRepository) CreateEventTx(ctx context.Context, tx pgx.Tx, event *model.WhatsAppGatewayEvent) (*model.WhatsAppGatewayEvent, error) {
	now := time.Now().UTC()
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = now
	}
	event.CreatedAt = now

	insertQuery := `
		INSERT INTO whatsapp_gateway_events (
			account_id, inbox_id, message_id, event_type, external_event_id,
			phone_number_id, business_account_id, raw_payload, received_at, processed_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (inbox_id, event_type, external_event_id)
		DO UPDATE SET processed_at = EXCLUDED.processed_at
		RETURNING id, account_id, inbox_id, message_id, event_type, external_event_id,
		          phone_number_id, business_account_id, raw_payload, received_at, processed_at, created_at
	`
	row := tx.QueryRow(
		ctx, insertQuery,
		event.AccountID, event.InboxID, event.MessageID, event.EventType, event.ExternalEventID,
		event.PhoneNumberID, event.BusinessAccountID, event.RawPayload, event.ReceivedAt, event.ProcessedAt, event.CreatedAt,
	)

	var res model.WhatsAppGatewayEvent
	err := row.Scan(
		&res.ID, &res.AccountID, &res.InboxID, &res.MessageID, &res.EventType, &res.ExternalEventID,
		&res.PhoneNumberID, &res.BusinessAccountID, &res.RawPayload, &res.ReceivedAt, &res.ProcessedAt, &res.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert gateway event in tx: %w", err)
	}
	return &res, nil
}

func (r *eventRepository) UpdateProcessed(ctx context.Context, eventID int64) error {
	now := time.Now().UTC()
	query := `UPDATE whatsapp_gateway_events SET processed_at = $1 WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, now, eventID)
	return err
}
