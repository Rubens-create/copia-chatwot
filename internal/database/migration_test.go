package database

import (
	"os"
	"regexp"
	"testing"
)

func TestAttachmentsExternalURLSupportsDataURLs(t *testing.T) {
	migration, err := os.ReadFile("../../migrations/001_chatwoot_whatsapp_schema.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	dataURLSafeColumn := regexp.MustCompile(`(?is)external_url\s+TEXT`)
	upgradeExistingColumn := regexp.MustCompile(`(?is)ALTER\s+TABLE\s+attachments\s+ALTER\s+COLUMN\s+external_url\s+TYPE\s+TEXT`)
	if !dataURLSafeColumn.Match(migration) {
		t.Fatal("attachments.external_url must be TEXT for image/audio Data URLs")
	}
	if !upgradeExistingColumn.Match(migration) {
		t.Fatal("migration must upgrade existing attachments.external_url columns to TEXT")
	}
}
