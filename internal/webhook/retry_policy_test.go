package webhook

import (
	"testing"
	"time"
)

func TestGetNextRetryDelay(t *testing.T) {
	tests := []struct {
		attempt       int
		expectedDelay time.Duration
		expectedRetry bool
	}{
		{attempt: 0, expectedDelay: 5 * time.Second, expectedRetry: true},
		{attempt: 1, expectedDelay: 5 * time.Second, expectedRetry: true},
		{attempt: 2, expectedDelay: 15 * time.Second, expectedRetry: true},
		{attempt: 3, expectedDelay: 1 * time.Minute, expectedRetry: true},
		{attempt: 4, expectedDelay: 5 * time.Minute, expectedRetry: true},
		{attempt: 5, expectedDelay: 0, expectedRetry: false}, // Exceeded -> dead letter
		{attempt: 6, expectedDelay: 0, expectedRetry: false},
	}

	for _, tt := range tests {
		delay, canRetry := GetNextRetryDelay(tt.attempt)
		if canRetry != tt.expectedRetry {
			t.Errorf("attempt %d: expected canRetry %v, got %v", tt.attempt, tt.expectedRetry, canRetry)
		}
		if canRetry && delay != tt.expectedDelay {
			t.Errorf("attempt %d: expected delay %v, got %v", tt.attempt, tt.expectedDelay, delay)
		}
	}
}
