package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/jackc/pgx/v5"
)

type MessageRepository interface {
	FindBySourceID(ctx context.Context, sourceID string) (*model.Message, error)
	FindBySourceIDTx(ctx context.Context, tx pgx.Tx, sourceID string) (*model.Message, error)
	Create(ctx context.Context, msg *model.Message) (*model.Message, error)
	CreateTx(ctx context.Context, tx pgx.Tx, msg *model.Message) (*model.Message, error)
	CreateAttachment(ctx context.Context, att *model.Attachment) (*model.Attachment, error)
	CreateAttachmentTx(ctx context.Context, tx pgx.Tx, att *model.Attachment) (*model.Attachment, error)
	UpdateStatusBySourceID(ctx context.Context, sourceID string, status int) error
	UpdateStatusBySourceIDTx(ctx context.Context, tx pgx.Tx, sourceID string, status int) error
	ListByConversationID(ctx context.Context, conversationID int, limit, offset int) ([]model.Message, error)
}

type messageRepository struct {
	db *database.DB
}

func NewMessageRepository(db *database.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) FindBySourceID(ctx context.Context, sourceID string) (*model.Message, error) {
	if sourceID == "" {
		return nil, nil
	}

	query := `
		SELECT id, content, account_id, inbox_id, conversation_id, message_type, content_type,
		       content_attributes, sender_type, sender_id, source_id, external_source_ids,
		       additional_attributes, processed_message_content, status, private, created_at, updated_at
		FROM messages
		WHERE source_id = $1
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, query, sourceID)

	var m model.Message
	err := row.Scan(
		&m.ID, &m.Content, &m.AccountID, &m.InboxID, &m.ConversationID, &m.MessageType, &m.ContentType,
		&m.ContentAttributes, &m.SenderType, &m.SenderID, &m.SourceID, &m.ExternalSourceIDs,
		&m.AdditionalAttributes, &m.ProcessedMessageContent, &m.Status, &m.Private, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding message by source_id: %w", err)
	}
	return &m, nil
}

func (r *messageRepository) FindBySourceIDTx(ctx context.Context, tx pgx.Tx, sourceID string) (*model.Message, error) {
	if sourceID == "" {
		return nil, nil
	}

	query := `
		SELECT id, content, account_id, inbox_id, conversation_id, message_type, content_type,
		       content_attributes, sender_type, sender_id, source_id, external_source_ids,
		       additional_attributes, processed_message_content, status, private, created_at, updated_at
		FROM messages
		WHERE source_id = $1
		LIMIT 1
	`
	row := tx.QueryRow(ctx, query, sourceID)

	var m model.Message
	err := row.Scan(
		&m.ID, &m.Content, &m.AccountID, &m.InboxID, &m.ConversationID, &m.MessageType, &m.ContentType,
		&m.ContentAttributes, &m.SenderType, &m.SenderID, &m.SourceID, &m.ExternalSourceIDs,
		&m.AdditionalAttributes, &m.ProcessedMessageContent, &m.Status, &m.Private, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding message by source_id in tx: %w", err)
	}
	return &m, nil
}

func (r *messageRepository) Create(ctx context.Context, msg *model.Message) (*model.Message, error) {
	now := time.Now().UTC()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	msg.UpdatedAt = now

	insertQuery := `
		INSERT INTO messages (
			content, account_id, inbox_id, conversation_id, message_type, content_type,
			content_attributes, sender_type, sender_id, source_id, external_source_ids,
			additional_attributes, processed_message_content, status, private, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (inbox_id, source_id) WHERE source_id IS NOT NULL
		DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id, content, account_id, inbox_id, conversation_id, message_type, content_type,
		          content_attributes, sender_type, sender_id, source_id, external_source_ids,
		          additional_attributes, processed_message_content, status, private, created_at, updated_at
	`

	contentAttrs := msg.ContentAttributes
	if len(contentAttrs) == 0 {
		contentAttrs = []byte("{}")
	}
	externalIDs := msg.ExternalSourceIDs
	if len(externalIDs) == 0 {
		externalIDs = []byte("{}")
	}
	additionalAttrs := msg.AdditionalAttributes
	if len(additionalAttrs) == 0 {
		additionalAttrs = []byte("{}")
	}

	row := r.db.Pool.QueryRow(
		ctx, insertQuery,
		msg.Content, msg.AccountID, msg.InboxID, msg.ConversationID, msg.MessageType, msg.ContentType,
		contentAttrs, msg.SenderType, msg.SenderID, msg.SourceID, externalIDs,
		additionalAttrs, msg.ProcessedMessageContent, msg.Status, msg.Private, msg.CreatedAt, msg.UpdatedAt,
	)

	var created model.Message
	err := row.Scan(
		&created.ID, &created.Content, &created.AccountID, &created.InboxID, &created.ConversationID,
		&created.MessageType, &created.ContentType, &created.ContentAttributes, &created.SenderType,
		&created.SenderID, &created.SourceID, &created.ExternalSourceIDs, &created.AdditionalAttributes,
		&created.ProcessedMessageContent, &created.Status, &created.Private, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %w", err)
	}

	return &created, nil
}

func (r *messageRepository) CreateTx(ctx context.Context, tx pgx.Tx, msg *model.Message) (*model.Message, error) {
	now := time.Now().UTC()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	msg.UpdatedAt = now

	insertQuery := `
		INSERT INTO messages (
			content, account_id, inbox_id, conversation_id, message_type, content_type,
			content_attributes, sender_type, sender_id, source_id, external_source_ids,
			additional_attributes, processed_message_content, status, private, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (inbox_id, source_id) WHERE source_id IS NOT NULL
		DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id, content, account_id, inbox_id, conversation_id, message_type, content_type,
		          content_attributes, sender_type, sender_id, source_id, external_source_ids,
		          additional_attributes, processed_message_content, status, private, created_at, updated_at
	`

	contentAttrs := msg.ContentAttributes
	if len(contentAttrs) == 0 {
		contentAttrs = []byte("{}")
	}
	externalIDs := msg.ExternalSourceIDs
	if len(externalIDs) == 0 {
		externalIDs = []byte("{}")
	}
	additionalAttrs := msg.AdditionalAttributes
	if len(additionalAttrs) == 0 {
		additionalAttrs = []byte("{}")
	}

	row := tx.QueryRow(
		ctx, insertQuery,
		msg.Content, msg.AccountID, msg.InboxID, msg.ConversationID, msg.MessageType, msg.ContentType,
		contentAttrs, msg.SenderType, msg.SenderID, msg.SourceID, externalIDs,
		additionalAttrs, msg.ProcessedMessageContent, msg.Status, msg.Private, msg.CreatedAt, msg.UpdatedAt,
	)

	var created model.Message
	err := row.Scan(
		&created.ID, &created.Content, &created.AccountID, &created.InboxID, &created.ConversationID,
		&created.MessageType, &created.ContentType, &created.ContentAttributes, &created.SenderType,
		&created.SenderID, &created.SourceID, &created.ExternalSourceIDs, &created.AdditionalAttributes,
		&created.ProcessedMessageContent, &created.Status, &created.Private, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert message in tx: %w", err)
	}

	return &created, nil
}

func (r *messageRepository) CreateAttachment(ctx context.Context, att *model.Attachment) (*model.Attachment, error) {
	now := time.Now().UTC()
	insertQuery := `
		INSERT INTO attachments (
			file_type, external_url, coordinates_lat, coordinates_long, message_id,
			account_id, fallback_title, extension, meta, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id, file_type, external_url, coordinates_lat, coordinates_long, message_id,
		          account_id, fallback_title, extension, meta, created_at, updated_at
	`

	meta := att.Meta
	if len(meta) == 0 {
		meta = []byte("{}")
	}

	row := r.db.Pool.QueryRow(
		ctx, insertQuery,
		att.FileType, att.ExternalURL, att.CoordinatesLat, att.CoordinatesLong, att.MessageID,
		att.AccountID, att.FallbackTitle, att.Extension, meta, now,
	)

	var created model.Attachment
	err := row.Scan(
		&created.ID, &created.FileType, &created.ExternalURL, &created.CoordinatesLat,
		&created.CoordinatesLong, &created.MessageID, &created.AccountID, &created.FallbackTitle,
		&created.Extension, &created.Meta, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert attachment: %w", err)
	}

	return &created, nil
}

func (r *messageRepository) CreateAttachmentTx(ctx context.Context, tx pgx.Tx, att *model.Attachment) (*model.Attachment, error) {
	now := time.Now().UTC()
	insertQuery := `
		INSERT INTO attachments (
			file_type, external_url, coordinates_lat, coordinates_long, message_id,
			account_id, fallback_title, extension, meta, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		RETURNING id, file_type, external_url, coordinates_lat, coordinates_long, message_id,
		          account_id, fallback_title, extension, meta, created_at, updated_at
	`

	meta := att.Meta
	if len(meta) == 0 {
		meta = []byte("{}")
	}

	row := tx.QueryRow(
		ctx, insertQuery,
		att.FileType, att.ExternalURL, att.CoordinatesLat, att.CoordinatesLong, att.MessageID,
		att.AccountID, att.FallbackTitle, att.Extension, meta, now,
	)

	var created model.Attachment
	err := row.Scan(
		&created.ID, &created.FileType, &created.ExternalURL, &created.CoordinatesLat,
		&created.CoordinatesLong, &created.MessageID, &created.AccountID, &created.FallbackTitle,
		&created.Extension, &created.Meta, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert attachment in tx: %w", err)
	}

	return &created, nil
}

func (r *messageRepository) UpdateStatusBySourceID(ctx context.Context, sourceID string, status int) error {
	now := time.Now().UTC()
	query := `UPDATE messages SET status = $1, updated_at = $2 WHERE source_id = $3`
	_, err := r.db.Pool.Exec(ctx, query, status, now, sourceID)
	return err
}

func (r *messageRepository) UpdateStatusBySourceIDTx(ctx context.Context, tx pgx.Tx, sourceID string, status int) error {
	now := time.Now().UTC()
	query := `UPDATE messages SET status = $1, updated_at = $2 WHERE source_id = $3`
	_, err := tx.Exec(ctx, query, status, now, sourceID)
	return err
}

func (r *messageRepository) ListByConversationID(ctx context.Context, conversationID int, limit, offset int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, content, account_id, inbox_id, conversation_id, message_type, content_type,
		       content_attributes, sender_type, sender_id, source_id, external_source_ids,
		       additional_attributes, processed_message_content, status, private, created_at, updated_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
	messageIDs := make([]int, 0)

	for rows.Next() {
		var m model.Message
		err := rows.Scan(
			&m.ID, &m.Content, &m.AccountID, &m.InboxID, &m.ConversationID, &m.MessageType, &m.ContentType,
			&m.ContentAttributes, &m.SenderType, &m.SenderID, &m.SourceID, &m.ExternalSourceIDs,
			&m.AdditionalAttributes, &m.ProcessedMessageContent, &m.Status, &m.Private, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		m.Attachments = make([]model.Attachment, 0)
		messages = append(messages, m)
		messageIDs = append(messageIDs, m.ID)
	}

	// Batch load attachments for these messages
	if len(messageIDs) > 0 {
		attQuery := `
			SELECT id, file_type, external_url, coordinates_lat, coordinates_long, message_id,
			       account_id, fallback_title, extension, meta, created_at, updated_at
			FROM attachments
			WHERE message_id = ANY($1)
		`
		attRows, err := r.db.Pool.Query(ctx, attQuery, messageIDs)
		if err == nil {
			defer attRows.Close()
			attMap := make(map[int][]model.Attachment)
			for attRows.Next() {
				var att model.Attachment
				if err := attRows.Scan(
					&att.ID, &att.FileType, &att.ExternalURL, &att.CoordinatesLat, &att.CoordinatesLong,
					&att.MessageID, &att.AccountID, &att.FallbackTitle, &att.Extension, &att.Meta,
					&att.CreatedAt, &att.UpdatedAt,
				); err == nil {
					attMap[att.MessageID] = append(attMap[att.MessageID], att)
				}
			}

			for i := range messages {
				if atts, found := attMap[messages[i].ID]; found {
					messages[i].Attachments = atts
				}
			}
		}
	}

	return messages, nil
}
