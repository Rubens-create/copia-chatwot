package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/chatwoot-lite/whatsapp-gateway/internal/database"
	"github.com/chatwoot-lite/whatsapp-gateway/internal/model"
	"github.com/jackc/pgx/v5"
)

type AccountRepository interface {
	FindByID(ctx context.Context, id int) (*model.Account, error)
	FindChannelByPhoneOrConfig(ctx context.Context, phone, phoneID string) (*model.ChannelWhatsapp, error)
	FindInboxByID(ctx context.Context, id int) (*model.Inbox, error)
	FindInboxByChannelID(ctx context.Context, channelID int) (*model.Inbox, error)
}

type accountRepository struct {
	db *database.DB
}

func NewAccountRepository(db *database.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) FindByID(ctx context.Context, id int) (*model.Account, error) {
	query := `SELECT id, name, locale, settings, created_at, updated_at FROM accounts WHERE id = $1`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var acc model.Account
	err := row.Scan(&acc.ID, &acc.Name, &acc.Locale, &acc.Settings, &acc.CreatedAt, &acc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding account: %w", err)
	}
	return &acc, nil
}

func (r *accountRepository) FindChannelByPhoneOrConfig(ctx context.Context, phone, phoneID string) (*model.ChannelWhatsapp, error) {
	query := `
		SELECT id, account_id, phone_number, provider, provider_config, business_management_token, created_at, updated_at
		FROM channel_whatsapp
		WHERE phone_number = $1 OR (provider_config->>'phone_number_id') = $2
		LIMIT 1
	`
	row := r.db.Pool.QueryRow(ctx, query, phone, phoneID)

	var ch model.ChannelWhatsapp
	err := row.Scan(&ch.ID, &ch.AccountID, &ch.PhoneNumber, &ch.Provider, &ch.ProviderConfig, &ch.BusinessManagementToken, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding channel_whatsapp: %w", err)
	}
	return &ch, nil
}

func (r *accountRepository) FindInboxByID(ctx context.Context, id int) (*model.Inbox, error) {
	query := `SELECT id, channel_id, account_id, name, channel_type, lock_to_single_conversation, created_at, updated_at FROM inboxes WHERE id = $1`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var inb model.Inbox
	err := row.Scan(&inb.ID, &inb.ChannelID, &inb.AccountID, &inb.Name, &inb.ChannelType, &inb.LockToSingleConversation, &inb.CreatedAt, &inb.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding inbox: %w", err)
	}
	return &inb, nil
}

func (r *accountRepository) FindInboxByChannelID(ctx context.Context, channelID int) (*model.Inbox, error) {
	query := `SELECT id, channel_id, account_id, name, channel_type, lock_to_single_conversation, created_at, updated_at FROM inboxes WHERE channel_id = $1 AND channel_type = 'Channel::Whatsapp' LIMIT 1`
	row := r.db.Pool.QueryRow(ctx, query, channelID)

	var inb model.Inbox
	err := row.Scan(&inb.ID, &inb.ChannelID, &inb.AccountID, &inb.Name, &inb.ChannelType, &inb.LockToSingleConversation, &inb.CreatedAt, &inb.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding inbox by channel_id: %w", err)
	}
	return &inb, nil
}
