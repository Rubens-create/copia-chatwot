package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ConversationRepository interface {
	FindOpenByContactAndInbox(ctx context.Context, accountID, inboxID int, contactID int64) (*model.Conversation, error)
	FindOpenByContactAndInboxTx(ctx context.Context, tx pgx.Tx, accountID, inboxID int, contactID int64) (*model.Conversation, error)
	FindLatestByContactAndInbox(ctx context.Context, accountID, inboxID int, contactID int64) (*model.Conversation, error)
	FindLatestByContactAndInboxTx(ctx context.Context, tx pgx.Tx, accountID, inboxID int, contactID int64) (*model.Conversation, error)
	Reopen(ctx context.Context, conversationID int) error
	ReopenTx(ctx context.Context, tx pgx.Tx, conversationID int) error
	Create(ctx context.Context, conv *model.Conversation) (*model.Conversation, error)
	CreateTx(ctx context.Context, tx pgx.Tx, conv *model.Conversation) (*model.Conversation, error)
	UpdateLastActivity(ctx context.Context, conversationID int) error
	UpdateLastActivityTx(ctx context.Context, tx pgx.Tx, conversationID int) error
	UpdateAdditionalAttributesTx(ctx context.Context, tx pgx.Tx, conversationID int, additionalAttrs []byte) error
	List(ctx context.Context, accountID, inboxID int, status *int, limit, offset int) ([]model.Conversation, error)
	FindByID(ctx context.Context, id int) (*model.Conversation, error)
}

type conversationRepository struct {
	db *database.DB
}

func NewConversationRepository(db *database.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) FindOpenByContactAndInbox(ctx context.Context, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	query := `
		SELECT id, account_id, inbox_id, contact_id, contact_inbox_id, status, display_id, uuid, identifier,
		       last_activity_at, additional_attributes, custom_attributes, created_at, updated_at
		FROM conversations
		WHERE account_id = $1 AND inbox_id = $2 AND contact_id = $3 AND status = $4
		ORDER BY last_activity_at DESC
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, query, accountID, inboxID, contactID, model.ConversationStatusOpen)

	var c model.Conversation
	err := row.Scan(
		&c.ID, &c.AccountID, &c.InboxID, &c.ContactID, &c.ContactInboxID, &c.Status, &c.DisplayID, &c.UUID, &c.Identifier,
		&c.LastActivityAt, &c.AdditionalAttributes, &c.CustomAttributes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding open conversation: %w", err)
	}
	return &c, nil
}

func (r *conversationRepository) FindOpenByContactAndInboxTx(ctx context.Context, tx pgx.Tx, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	query := `
		SELECT id, account_id, inbox_id, contact_id, contact_inbox_id, status, display_id, uuid, identifier,
		       last_activity_at, additional_attributes, custom_attributes, created_at, updated_at
		FROM conversations
		WHERE account_id = $1 AND inbox_id = $2 AND contact_id = $3 AND status = $4
		ORDER BY last_activity_at DESC
		LIMIT 1
	`
	row := tx.QueryRow(ctx, query, accountID, inboxID, contactID, model.ConversationStatusOpen)

	var c model.Conversation
	err := row.Scan(
		&c.ID, &c.AccountID, &c.InboxID, &c.ContactID, &c.ContactInboxID, &c.Status, &c.DisplayID, &c.UUID, &c.Identifier,
		&c.LastActivityAt, &c.AdditionalAttributes, &c.CustomAttributes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding open conversation in tx: %w", err)
	}
	return &c, nil
}

func (r *conversationRepository) FindLatestByContactAndInbox(ctx context.Context, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	query := `
		SELECT id, account_id, inbox_id, contact_id, contact_inbox_id, status, display_id, uuid, identifier,
		       last_activity_at, additional_attributes, custom_attributes, created_at, updated_at
		FROM conversations
		WHERE account_id = $1 AND inbox_id = $2 AND contact_id = $3
		ORDER BY last_activity_at DESC
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, query, accountID, inboxID, contactID)

	var c model.Conversation
	err := row.Scan(
		&c.ID, &c.AccountID, &c.InboxID, &c.ContactID, &c.ContactInboxID, &c.Status, &c.DisplayID, &c.UUID, &c.Identifier,
		&c.LastActivityAt, &c.AdditionalAttributes, &c.CustomAttributes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding latest conversation: %w", err)
	}
	return &c, nil
}

func (r *conversationRepository) FindLatestByContactAndInboxTx(ctx context.Context, tx pgx.Tx, accountID, inboxID int, contactID int64) (*model.Conversation, error) {
	query := `
		SELECT id, account_id, inbox_id, contact_id, contact_inbox_id, status, display_id, uuid, identifier,
		       last_activity_at, additional_attributes, custom_attributes, created_at, updated_at
		FROM conversations
		WHERE account_id = $1 AND inbox_id = $2 AND contact_id = $3
		ORDER BY last_activity_at DESC
		LIMIT 1
	`
	row := tx.QueryRow(ctx, query, accountID, inboxID, contactID)

	var c model.Conversation
	err := row.Scan(
		&c.ID, &c.AccountID, &c.InboxID, &c.ContactID, &c.ContactInboxID, &c.Status, &c.DisplayID, &c.UUID, &c.Identifier,
		&c.LastActivityAt, &c.AdditionalAttributes, &c.CustomAttributes, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding latest conversation in tx: %w", err)
	}
	return &c, nil
}

func (r *conversationRepository) Reopen(ctx context.Context, conversationID int) error {
	now := time.Now().UTC()
	query := `UPDATE conversations SET status = 0, last_activity_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, now, conversationID)
	return err
}

func (r *conversationRepository) ReopenTx(ctx context.Context, tx pgx.Tx, conversationID int) error {
	now := time.Now().UTC()
	query := `UPDATE conversations SET status = 0, last_activity_at = $1, updated_at = $1 WHERE id = $2`
	_, err := tx.Exec(ctx, query, now, conversationID)
	return err
}

func (r *conversationRepository) Create(ctx context.Context, conv *model.Conversation) (*model.Conversation, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	created, err := r.CreateTx(ctx, tx, conv)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit conversation transaction: %w", err)
	}

	return created, nil
}

func (r *conversationRepository) CreateTx(ctx context.Context, tx pgx.Tx, conv *model.Conversation) (*model.Conversation, error) {
	now := time.Now().UTC()
	if conv.UUID == "" {
		conv.UUID = uuid.New().String()
	}

	// 1. Transactional advisory lock per account to eliminate display_id race conditions under high concurrency
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, conv.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire advisory lock for display_id: %w", err)
	}

	// 2. Get next display_id for account safely
	var nextDisplayID int
	err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(display_id), 0) + 1 FROM conversations WHERE account_id = $1`, conv.AccountID).Scan(&nextDisplayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next display_id: %w", err)
	}
	conv.DisplayID = nextDisplayID

	insertQuery := `
		INSERT INTO conversations (
			account_id, inbox_id, contact_id, contact_inbox_id, status, display_id, uuid, identifier,
			last_activity_at, additional_attributes, custom_attributes, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $9, $9)
		RETURNING id, account_id, inbox_id, contact_id, contact_inbox_id, status, display_id, uuid, identifier,
		          last_activity_at, additional_attributes, custom_attributes, created_at, updated_at
	`
	additionalAttrs := conv.AdditionalAttributes
	if len(additionalAttrs) == 0 {
		additionalAttrs = []byte("{}")
	}
	customAttrs := conv.CustomAttributes
	if len(customAttrs) == 0 {
		customAttrs = []byte("{}")
	}

	row := tx.QueryRow(
		ctx, insertQuery,
		conv.AccountID, conv.InboxID, conv.ContactID, conv.ContactInboxID, conv.Status, conv.DisplayID, conv.UUID, conv.Identifier,
		now, additionalAttrs, customAttrs,
	)

	var created model.Conversation
	err = row.Scan(
		&created.ID, &created.AccountID, &created.InboxID, &created.ContactID, &created.ContactInboxID, &created.Status,
		&created.DisplayID, &created.UUID, &created.Identifier, &created.LastActivityAt, &created.AdditionalAttributes,
		&created.CustomAttributes, &created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert conversation: %w", err)
	}

	return &created, nil
}

func (r *conversationRepository) UpdateLastActivity(ctx context.Context, conversationID int) error {
	now := time.Now().UTC()
	query := `UPDATE conversations SET last_activity_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, now, conversationID)
	return err
}

func (r *conversationRepository) UpdateLastActivityTx(ctx context.Context, tx pgx.Tx, conversationID int) error {
	query := `UPDATE conversations SET last_activity_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := tx.Exec(ctx, query, conversationID)
	return err
}

func (r *conversationRepository) UpdateAdditionalAttributesTx(ctx context.Context, tx pgx.Tx, conversationID int, additionalAttrs []byte) error {
	query := `UPDATE conversations SET additional_attributes = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := tx.Exec(ctx, query, additionalAttrs, conversationID)
	return err
}

func (r *conversationRepository) List(ctx context.Context, accountID, inboxID int, status *int, limit, offset int) ([]model.Conversation, error) {
	if limit <= 0 {
		limit = 25
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT c.id, c.account_id, c.inbox_id, c.contact_id, c.contact_inbox_id, c.status, c.display_id, c.uuid, c.identifier,
		       c.last_activity_at, c.additional_attributes, c.custom_attributes, c.created_at, c.updated_at,
		       ct.id, ct.name, ct.email, ct.phone_number, ct.account_id, ct.identifier, ct.additional_attributes, ct.custom_attributes, ct.last_activity_at, ct.created_at, ct.updated_at
		FROM conversations c
		JOIN contacts ct ON c.contact_id = ct.id
		WHERE c.account_id = $1 AND ($2 = 0 OR c.inbox_id = $2) AND ($3::integer IS NULL OR c.status = $3)
		ORDER BY c.last_activity_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.Pool.Query(ctx, query, accountID, inboxID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]model.Conversation, 0)
	for rows.Next() {
		var c model.Conversation
		var ct model.Contact
		err := rows.Scan(
			&c.ID, &c.AccountID, &c.InboxID, &c.ContactID, &c.ContactInboxID, &c.Status, &c.DisplayID, &c.UUID, &c.Identifier,
			&c.LastActivityAt, &c.AdditionalAttributes, &c.CustomAttributes, &c.CreatedAt, &c.UpdatedAt,
			&ct.ID, &ct.Name, &ct.Email, &ct.PhoneNumber, &ct.AccountID, &ct.Identifier, &ct.AdditionalAttributes, &ct.CustomAttributes, &ct.LastActivityAt, &ct.CreatedAt, &ct.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		c.Contact = &ct

		// Load last message for preview
		lastMsgQuery := `
			SELECT id, content, message_type, content_type, status, created_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		`
		var lastMsg model.Message
		err = r.db.Pool.QueryRow(ctx, lastMsgQuery, c.ID).Scan(
			&lastMsg.ID, &lastMsg.Content, &lastMsg.MessageType, &lastMsg.ContentType, &lastMsg.Status, &lastMsg.CreatedAt,
		)
		if err == nil {
			c.LastMessage = &lastMsg
		}

		conversations = append(conversations, c)
	}

	return conversations, nil
}

func (r *conversationRepository) FindByID(ctx context.Context, id int) (*model.Conversation, error) {
	query := `
		SELECT c.id, c.account_id, c.inbox_id, c.contact_id, c.contact_inbox_id, c.status, c.display_id, c.uuid, c.identifier,
		       c.last_activity_at, c.additional_attributes, c.custom_attributes, c.created_at, c.updated_at,
		       ct.id, ct.name, ct.email, ct.phone_number, ct.account_id, ct.identifier, ct.additional_attributes, ct.custom_attributes, ct.last_activity_at, ct.created_at, ct.updated_at
		FROM conversations c
		JOIN contacts ct ON c.contact_id = ct.id
		WHERE c.id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var c model.Conversation
	var ct model.Contact
	err := row.Scan(
		&c.ID, &c.AccountID, &c.InboxID, &c.ContactID, &c.ContactInboxID, &c.Status, &c.DisplayID, &c.UUID, &c.Identifier,
		&c.LastActivityAt, &c.AdditionalAttributes, &c.CustomAttributes, &c.CreatedAt, &c.UpdatedAt,
		&ct.ID, &ct.Name, &ct.Email, &ct.PhoneNumber, &ct.AccountID, &ct.Identifier, &ct.AdditionalAttributes, &ct.CustomAttributes, &ct.LastActivityAt, &ct.CreatedAt, &ct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding conversation by id: %w", err)
	}
	c.Contact = &ct
	return &c, nil
}
