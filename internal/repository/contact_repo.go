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

type ContactRepository interface {
	FindOrCreateByPhone(ctx context.Context, accountID int, phone, name string) (*model.Contact, error)
	FindOrCreateByPhoneTx(ctx context.Context, tx pgx.Tx, accountID int, phone, name string) (*model.Contact, error)
	FindOrCreateContactInbox(ctx context.Context, contactID, inboxID int64, sourceID string) (*model.ContactInbox, error)
	FindOrCreateContactInboxTx(ctx context.Context, tx pgx.Tx, contactID, inboxID int64, sourceID string) (*model.ContactInbox, error)
	FindByID(ctx context.Context, id int) (*model.Contact, error)
}

type contactRepository struct {
	db *database.DB
}

func NewContactRepository(db *database.DB) ContactRepository {
	return &contactRepository{db: db}
}

func (r *contactRepository) FindOrCreateByPhone(ctx context.Context, accountID int, phone, name string) (*model.Contact, error) {
	now := time.Now().UTC()

	// Try to find existing contact by phone and account_id
	queryFind := `
		SELECT id, name, email, phone_number, account_id, identifier, additional_attributes, custom_attributes, last_activity_at, created_at, updated_at
		FROM contacts
		WHERE account_id = $1 AND phone_number = $2
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, queryFind, accountID, phone)

	var c model.Contact
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.PhoneNumber, &c.AccountID, &c.Identifier, &c.AdditionalAttributes, &c.CustomAttributes, &c.LastActivityAt, &c.CreatedAt, &c.UpdatedAt)
	if err == nil {
		// Existing contact found, update activity and name if appropriate
		if name != "" && (c.Name == "" || c.Name == phone) {
			updateQuery := `UPDATE contacts SET name = $1, last_activity_at = $2, updated_at = $2 WHERE id = $3`
			_, _ = r.db.Pool.Exec(ctx, updateQuery, name, now, c.ID)
			c.Name = name
		} else {
			updateQuery := `UPDATE contacts SET last_activity_at = $1, updated_at = $1 WHERE id = $2`
			_, _ = r.db.Pool.Exec(ctx, updateQuery, now, c.ID)
		}
		c.LastActivityAt = &now
		return &c, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error finding contact: %w", err)
	}

	// Create new contact
	if name == "" {
		name = phone
	}

	insertQuery := `
		INSERT INTO contacts (name, phone_number, account_id, created_at, updated_at, last_activity_at, additional_attributes, custom_attributes)
		VALUES ($1, $2, $3, $4, $4, $4, '{}'::jsonb, '{}'::jsonb)
		RETURNING id, name, email, phone_number, account_id, identifier, additional_attributes, custom_attributes, last_activity_at, created_at, updated_at
	`
	row = r.db.Pool.QueryRow(ctx, insertQuery, name, phone, accountID, now)

	err = row.Scan(&c.ID, &c.Name, &c.Email, &c.PhoneNumber, &c.AccountID, &c.Identifier, &c.AdditionalAttributes, &c.CustomAttributes, &c.LastActivityAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting contact: %w", err)
	}

	return &c, nil
}

func (r *contactRepository) FindOrCreateByPhoneTx(ctx context.Context, tx pgx.Tx, accountID int, phone, name string) (*model.Contact, error) {
	now := time.Now().UTC()

	queryFind := `
		SELECT id, name, email, phone_number, account_id, identifier, additional_attributes, custom_attributes, last_activity_at, created_at, updated_at
		FROM contacts
		WHERE account_id = $1 AND phone_number = $2
		LIMIT 1
	`
	row := tx.QueryRow(ctx, queryFind, accountID, phone)

	var c model.Contact
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.PhoneNumber, &c.AccountID, &c.Identifier, &c.AdditionalAttributes, &c.CustomAttributes, &c.LastActivityAt, &c.CreatedAt, &c.UpdatedAt)
	if err == nil {
		if name != "" && (c.Name == "" || c.Name == phone) {
			updateQuery := `UPDATE contacts SET name = $1, last_activity_at = $2, updated_at = $2 WHERE id = $3`
			_, _ = tx.Exec(ctx, updateQuery, name, now, c.ID)
			c.Name = name
		} else {
			updateQuery := `UPDATE contacts SET last_activity_at = $1, updated_at = $1 WHERE id = $2`
			_, _ = tx.Exec(ctx, updateQuery, now, c.ID)
		}
		c.LastActivityAt = &now
		return &c, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error finding contact: %w", err)
	}

	if name == "" {
		name = phone
	}

	insertQuery := `
		INSERT INTO contacts (name, phone_number, account_id, created_at, updated_at, last_activity_at, additional_attributes, custom_attributes)
		VALUES ($1, $2, $3, $4, $4, $4, '{}'::jsonb, '{}'::jsonb)
		RETURNING id, name, email, phone_number, account_id, identifier, additional_attributes, custom_attributes, last_activity_at, created_at, updated_at
	`
	row = tx.QueryRow(ctx, insertQuery, name, phone, accountID, now)
	err = row.Scan(&c.ID, &c.Name, &c.Email, &c.PhoneNumber, &c.AccountID, &c.Identifier, &c.AdditionalAttributes, &c.CustomAttributes, &c.LastActivityAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting contact in tx: %w", err)
	}

	return &c, nil
}

func (r *contactRepository) FindOrCreateContactInbox(ctx context.Context, contactID, inboxID int64, sourceID string) (*model.ContactInbox, error) {
	now := time.Now().UTC()

	queryFind := `
		SELECT id, contact_id, inbox_id, source_id, created_at, updated_at
		FROM contact_inboxes
		WHERE inbox_id = $1 AND source_id = $2
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, queryFind, inboxID, sourceID)

	var ci model.ContactInbox
	err := row.Scan(&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt)
	if err == nil {
		return &ci, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error finding contact_inbox: %w", err)
	}

	insertQuery := `
		INSERT INTO contact_inboxes (contact_id, inbox_id, source_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (inbox_id, source_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id, contact_id, inbox_id, source_id, created_at, updated_at
	`
	row = r.db.Pool.QueryRow(ctx, insertQuery, contactID, inboxID, sourceID, now)
	err = row.Scan(&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting contact_inbox: %w", err)
	}

	return &ci, nil
}

func (r *contactRepository) FindOrCreateContactInboxTx(ctx context.Context, tx pgx.Tx, contactID, inboxID int64, sourceID string) (*model.ContactInbox, error) {
	now := time.Now().UTC()

	queryFind := `
		SELECT id, contact_id, inbox_id, source_id, created_at, updated_at
		FROM contact_inboxes
		WHERE inbox_id = $1 AND source_id = $2
		LIMIT 1
	`
	row := tx.QueryRow(ctx, queryFind, inboxID, sourceID)

	var ci model.ContactInbox
	err := row.Scan(&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt)
	if err == nil {
		return &ci, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error finding contact_inbox: %w", err)
	}

	insertQuery := `
		INSERT INTO contact_inboxes (contact_id, inbox_id, source_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (inbox_id, source_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id, contact_id, inbox_id, source_id, created_at, updated_at
	`
	row = tx.QueryRow(ctx, insertQuery, contactID, inboxID, sourceID, now)
	err = row.Scan(&ci.ID, &ci.ContactID, &ci.InboxID, &ci.SourceID, &ci.CreatedAt, &ci.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("error inserting contact_inbox in tx: %w", err)
	}

	return &ci, nil
}

func (r *contactRepository) FindByID(ctx context.Context, id int) (*model.Contact, error) {
	query := `
		SELECT id, name, email, phone_number, account_id, identifier, additional_attributes, custom_attributes, last_activity_at, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var c model.Contact
	err := row.Scan(&c.ID, &c.Name, &c.Email, &c.PhoneNumber, &c.AccountID, &c.Identifier, &c.AdditionalAttributes, &c.CustomAttributes, &c.LastActivityAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding contact by id: %w", err)
	}
	return &c, nil
}
