package whatsapp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type captureRoundTripper struct {
	payload map[string]interface{}
}

func (c *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &c.payload); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(`{"messages":[{"id":"wamid.test"}]}`)),
	}, nil
}

func TestSendAudioMarksRecordedAudioAsVoiceMessage(t *testing.T) {
	transport := &captureRoundTripper{}
	client := &metaClient{
		accessToken: "test-token",
		apiVersion:  "v23.0",
		httpClient:  &http.Client{Transport: transport},
	}

	if _, err := client.SendAudio(context.Background(), "phone-id", "+5511999999999", "media-id", true); err != nil {
		t.Fatalf("SendAudio returned error: %v", err)
	}

	audio, ok := transport.payload["audio"].(map[string]interface{})
	if !ok {
		t.Fatalf("audio payload missing: %#v", transport.payload)
	}
	if got := audio["voice"]; got != true {
		t.Fatalf("audio.voice = %#v; want true", got)
	}
}

func TestSendAudioOmitsVoiceForBasicAudio(t *testing.T) {
	transport := &captureRoundTripper{}
	client := &metaClient{
		accessToken: "test-token",
		apiVersion:  "v23.0",
		httpClient:  &http.Client{Transport: transport},
	}

	if _, err := client.SendAudio(context.Background(), "phone-id", "+5511999999999", "media-id", false); err != nil {
		t.Fatalf("SendAudio returned error: %v", err)
	}

	audio, ok := transport.payload["audio"].(map[string]interface{})
	if !ok {
		t.Fatalf("audio payload missing: %#v", transport.payload)
	}
	if _, exists := audio["voice"]; exists {
		t.Fatalf("audio.voice should be omitted for basic audio: %#v", audio)
	}
}
