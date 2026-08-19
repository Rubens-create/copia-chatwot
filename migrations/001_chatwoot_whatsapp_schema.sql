-- Chatwoot-Compatible PostgreSQL Schema for WhatsApp Gateway
-- Uses CREATE TABLE IF NOT EXISTS to guarantee 100% compatibility with existing Chatwoot databases.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "plpgsql";

-- 1. Accounts
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    locale INTEGER DEFAULT 0,
    domain VARCHAR(100),
    support_email VARCHAR(100),
    feature_flags BIGINT DEFAULT 0 NOT NULL,
    auto_resolve_duration INTEGER,
    limits JSONB DEFAULT '{}'::jsonb,
    custom_attributes JSONB DEFAULT '{}'::jsonb,
    status INTEGER DEFAULT 0,
    internal_attributes JSONB DEFAULT '{}'::jsonb NOT NULL,
    settings JSONB DEFAULT '{}'::jsonb,
    feature_flags_ext_1 BIGINT DEFAULT 0 NOT NULL
);

-- 2. Channel WhatsApp
CREATE TABLE IF NOT EXISTS channel_whatsapp (
    id BIGSERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    business_management_token TEXT,
    phone_number VARCHAR(255) NOT NULL,
    provider VARCHAR(255) DEFAULT 'default',
    provider_config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message_templates JSONB DEFAULT '{}'::jsonb,
    message_templates_last_updated TIMESTAMP WITHOUT TIME ZONE,
    phone_number_health JSONB DEFAULT '{}'::jsonb NOT NULL,
    phone_number_health_checked_at TIMESTAMP WITHOUT TIME ZONE,
    phone_number_health_error VARCHAR(500)
);

CREATE UNIQUE INDEX IF NOT EXISTS index_channel_whatsapp_on_phone_number ON channel_whatsapp (phone_number);

-- 3. Inboxes
CREATE TABLE IF NOT EXISTS inboxes (
    id SERIAL PRIMARY KEY,
    channel_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    channel_type VARCHAR(255) DEFAULT 'Channel::Whatsapp',
    enable_auto_assignment BOOLEAN DEFAULT TRUE,
    greeting_enabled BOOLEAN DEFAULT FALSE,
    greeting_message VARCHAR(255),
    email_address VARCHAR(255),
    working_hours_enabled BOOLEAN DEFAULT FALSE,
    out_of_office_message VARCHAR(255),
    timezone VARCHAR(255) DEFAULT 'UTC',
    enable_email_collect BOOLEAN DEFAULT TRUE,
    csat_survey_enabled BOOLEAN DEFAULT FALSE,
    allow_messages_after_resolved BOOLEAN DEFAULT TRUE,
    auto_assignment_config JSONB DEFAULT '{}'::jsonb,
    lock_to_single_conversation BOOLEAN DEFAULT FALSE NOT NULL,
    portal_id BIGINT,
    sender_name_type INTEGER DEFAULT 0 NOT NULL,
    business_name VARCHAR(255),
    csat_config JSONB DEFAULT '{}'::jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS index_inboxes_on_account_id ON inboxes (account_id);
CREATE INDEX IF NOT EXISTS index_inboxes_on_channel_id_and_channel_type ON inboxes (channel_id, channel_type);

-- 4. Contacts
CREATE TABLE IF NOT EXISTS contacts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) DEFAULT '',
    email VARCHAR(255),
    phone_number VARCHAR(255),
    account_id INTEGER NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    additional_attributes JSONB DEFAULT '{}'::jsonb,
    identifier VARCHAR(255),
    custom_attributes JSONB DEFAULT '{}'::jsonb,
    last_activity_at TIMESTAMP WITHOUT TIME ZONE,
    contact_type INTEGER DEFAULT 0,
    middle_name VARCHAR(255) DEFAULT '',
    last_name VARCHAR(255) DEFAULT '',
    location VARCHAR(255) DEFAULT '',
    country_code VARCHAR(255) DEFAULT '',
    blocked BOOLEAN DEFAULT FALSE NOT NULL,
    company_id BIGINT
);

CREATE INDEX IF NOT EXISTS index_contacts_on_account_id ON contacts (account_id);
CREATE INDEX IF NOT EXISTS index_contacts_on_phone_number_and_account_id ON contacts (phone_number, account_id);

-- 5. Contact Inboxes
CREATE TABLE IF NOT EXISTS contact_inboxes (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL,
    inbox_id BIGINT NOT NULL,
    source_id TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    hmac_verified BOOLEAN DEFAULT FALSE,
    pubsub_token VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS index_contact_inboxes_on_contact_id ON contact_inboxes (contact_id);
CREATE INDEX IF NOT EXISTS index_contact_inboxes_on_inbox_id ON contact_inboxes (inbox_id);
CREATE UNIQUE INDEX IF NOT EXISTS index_contact_inboxes_on_inbox_id_and_source_id ON contact_inboxes (inbox_id, source_id);

-- 6. Conversations
CREATE TABLE IF NOT EXISTS conversations (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    inbox_id INTEGER NOT NULL,
    status INTEGER DEFAULT 0 NOT NULL,
    assignee_id INTEGER,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    contact_id BIGINT,
    display_id INTEGER NOT NULL,
    contact_last_seen_at TIMESTAMP WITHOUT TIME ZONE,
    agent_last_seen_at TIMESTAMP WITHOUT TIME ZONE,
    additional_attributes JSONB DEFAULT '{}'::jsonb,
    contact_inbox_id BIGINT,
    uuid UUID DEFAULT gen_random_uuid() NOT NULL,
    identifier VARCHAR(255),
    last_activity_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    team_id BIGINT,
    campaign_id BIGINT,
    snoozed_until TIMESTAMP WITHOUT TIME ZONE,
    custom_attributes JSONB DEFAULT '{}'::jsonb,
    assignee_last_seen_at TIMESTAMP WITHOUT TIME ZONE,
    first_reply_created_at TIMESTAMP WITHOUT TIME ZONE,
    priority INTEGER,
    sla_policy_id BIGINT,
    waiting_since TIMESTAMP WITHOUT TIME ZONE,
    cached_label_list TEXT,
    assignee_agent_bot_id BIGINT,
    status_changed_at TIMESTAMP WITHOUT TIME ZONE
);

CREATE INDEX IF NOT EXISTS index_conversations_on_account_id ON conversations (account_id);
CREATE INDEX IF NOT EXISTS index_conversations_on_inbox_id ON conversations (inbox_id);
CREATE INDEX IF NOT EXISTS index_conversations_on_contact_id ON conversations (contact_id);
CREATE INDEX IF NOT EXISTS index_conversations_on_contact_inbox_id ON conversations (contact_inbox_id);
CREATE UNIQUE INDEX IF NOT EXISTS index_conversations_on_account_id_and_display_id ON conversations (account_id, display_id);
CREATE UNIQUE INDEX IF NOT EXISTS index_conversations_on_uuid ON conversations (uuid);
CREATE INDEX IF NOT EXISTS index_conversations_on_last_activity_at ON conversations (last_activity_at DESC);

-- 7. Messages
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    content TEXT,
    account_id INTEGER NOT NULL,
    inbox_id INTEGER NOT NULL,
    conversation_id INTEGER NOT NULL,
    message_type INTEGER NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    private BOOLEAN DEFAULT FALSE NOT NULL,
    status INTEGER DEFAULT 0,
    source_id TEXT,
    content_type INTEGER DEFAULT 0 NOT NULL,
    content_attributes JSONB DEFAULT '{}'::jsonb,
    sender_type VARCHAR(255),
    sender_id BIGINT,
    external_source_ids JSONB DEFAULT '{}'::jsonb,
    additional_attributes JSONB DEFAULT '{}'::jsonb,
    processed_message_content TEXT,
    sentiment JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS index_messages_on_account_id ON messages (account_id);
CREATE INDEX IF NOT EXISTS index_messages_on_inbox_id ON messages (inbox_id);
CREATE INDEX IF NOT EXISTS index_messages_on_conversation_id ON messages (conversation_id);
CREATE INDEX IF NOT EXISTS index_messages_on_source_id ON messages (source_id);
CREATE INDEX IF NOT EXISTS index_messages_on_created_at ON messages (created_at);

-- Partial Unique Index for Strict Message Idempotency per Inbox
CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_whatsapp_source_id_unique
ON messages (inbox_id, source_id)
WHERE source_id IS NOT NULL;

-- 8. Attachments
CREATE TABLE IF NOT EXISTS attachments (
    id SERIAL PRIMARY KEY,
    file_type INTEGER DEFAULT 0,
    external_url TEXT,
    coordinates_lat FLOAT DEFAULT 0.0,
    coordinates_long FLOAT DEFAULT 0.0,
    message_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fallback_title VARCHAR(255),
    extension VARCHAR(255),
    meta JSONB DEFAULT '{}'::jsonb
);

-- Existing installations used VARCHAR(255), which is too small for Base64
-- Data URLs used to preview outgoing images, audio and documents in the chat.
ALTER TABLE attachments
    ALTER COLUMN external_url TYPE TEXT;

CREATE INDEX IF NOT EXISTS index_attachments_on_account_id ON attachments (account_id);
CREATE INDEX IF NOT EXISTS index_attachments_on_message_id ON attachments (message_id);

-- 9. Webhooks
CREATE TABLE IF NOT EXISTS webhooks (
    id BIGSERIAL PRIMARY KEY,
    account_id INTEGER,
    inbox_id INTEGER,
    url TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    webhook_type INTEGER DEFAULT 0,
    subscriptions JSONB DEFAULT '["message_created", "message_updated", "conversation_created", "conversation_updated"]'::jsonb,
    name VARCHAR(255),
    secret VARCHAR(255)
);

CREATE UNIQUE INDEX IF NOT EXISTS index_webhooks_on_account_id_and_url ON webhooks (account_id, url);

-- 10. Webhook Delivery Attempts (Auxiliary table for reliable retries and dead-letter monitoring)
CREATE TABLE IF NOT EXISTS webhook_delivery_attempts (
    id BIGSERIAL PRIMARY KEY,
    webhook_id BIGINT REFERENCES webhooks(id) ON DELETE CASCADE,
    event VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    url TEXT NOT NULL,
    attempts INTEGER DEFAULT 0 NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' NOT NULL, -- pending, success, failed, dead_letter
    last_error TEXT,
    response_code INTEGER,
    last_attempt_at TIMESTAMP WITHOUT TIME ZONE,
    next_attempt_at TIMESTAMP WITHOUT TIME ZONE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS index_webhook_delivery_attempts_status ON webhook_delivery_attempts (status, next_attempt_at);

-- 11. WhatsApp Gateway Events Audit Table (Non-conflicting auxiliary log for raw Meta payloads & status updates)
CREATE TABLE IF NOT EXISTS whatsapp_gateway_events (
    id BIGSERIAL PRIMARY KEY,
    account_id INTEGER REFERENCES accounts(id),
    inbox_id INTEGER REFERENCES inboxes(id),
    message_id INTEGER REFERENCES messages(id),
    event_type VARCHAR(255) NOT NULL,
    external_event_id TEXT NOT NULL,
    phone_number_id VARCHAR(255),
    business_account_id VARCHAR(255),
    raw_payload JSONB NOT NULL,
    received_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    processed_at TIMESTAMP WITHOUT TIME ZONE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uniq_gateway_event UNIQUE(inbox_id, event_type, external_event_id)
);

CREATE INDEX IF NOT EXISTS index_whatsapp_gateway_events_on_external_id ON whatsapp_gateway_events (external_event_id);

-- 12. Initial Seeds (executed only if empty)
INSERT INTO accounts (id, name, locale, settings, created_at, updated_at)
VALUES (1, 'Default Account', 0, '{"timezone": "UTC"}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

INSERT INTO channel_whatsapp (id, account_id, phone_number, provider, provider_config, phone_number_health, created_at, updated_at)
VALUES (1, 1, '+15550000000', 'whatsapp_cloud', '{"phone_number_id": "default", "business_account_id": "default"}'::jsonb, '{}'::jsonb, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

INSERT INTO inboxes (id, channel_id, account_id, name, channel_type, created_at, updated_at)
VALUES (1, 1, 1, 'WhatsApp Inbox', 'Channel::Whatsapp', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

SELECT setval('accounts_id_seq', (SELECT GREATEST(COALESCE(MAX(id), 1), 1) FROM accounts));
SELECT setval('channel_whatsapp_id_seq', (SELECT GREATEST(COALESCE(MAX(id), 1), 1) FROM channel_whatsapp));
SELECT setval('inboxes_id_seq', (SELECT GREATEST(COALESCE(MAX(id), 1), 1) FROM inboxes));
