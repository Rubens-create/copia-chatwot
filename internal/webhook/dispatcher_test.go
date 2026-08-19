package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpDispatcher_Dispatch(t *testing.T) {
	secret := "test_secret_key"
	event := "message_created"
	payload := []byte(`{"event":"message_created","message":{"id":123,"content":"hello"}}`)

	var receivedEvent string
	var receivedSignature string
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEvent = r.Header.Get("X-Chatwoot-Event")
		receivedSignature = r.Header.Get("X-Chatwoot-Signature-256")
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer server.Close()

	dispatcher := NewDispatcher()
	status, err := dispatcher.Dispatch(context.Background(), server.URL, secret, event, payload)
	if err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	if status != http.StatusOK {
		t.Errorf("expected status 200, got %d", status)
	}

	if receivedEvent != event {
		t.Errorf("expected header X-Chatwoot-Event %q, got %q", event, receivedEvent)
	}

	if receivedBody != string(payload) {
		t.Errorf("expected body %q, got %q", string(payload), receivedBody)
	}

	// Verify HMAC-SHA256 signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if receivedSignature != expectedSig {
		t.Errorf("expected signature %q, got %q", expectedSig, receivedSignature)
	}
}
