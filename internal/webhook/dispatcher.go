package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, url, secret, event string, payload []byte) (int, error)
}

type httpDispatcher struct {
	client *http.Client
}

func NewDispatcher() WebhookDispatcher {
	return &httpDispatcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *httpDispatcher) Dispatch(ctx context.Context, url, secret, event string, payload []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Chatwoot-Lite-WhatsApp-Gateway/1.0")
	req.Header.Set("X-Chatwoot-Event", event)

	// If secret is configured, generate HMAC-SHA256 signature
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Chatwoot-Signature-256", "sha256="+signature)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http dispatch error: %w", err)
	}
	defer resp.Body.Close()

	// Drain body up to 4KB for error message
	bodyPreview, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("webhook endpoint returned status %d: %s", resp.StatusCode, string(bodyPreview))
	}

	return resp.StatusCode, nil
}
