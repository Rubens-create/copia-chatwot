package webhook

import (
	"time"
)

var retryDelays = []time.Duration{
	5 * time.Second,  // after attempt 1
	15 * time.Second, // after attempt 2
	1 * time.Minute,  // after attempt 3
	5 * time.Minute,  // after attempt 4
}

const MaxRetryAttempts = 4

// GetNextRetryDelay returns the delay for the next attempt.
// If attempts exceed MaxRetryAttempts, returns 0 and false (indicating dead-letter).
func GetNextRetryDelay(attempt int) (time.Duration, bool) {
	if attempt <= 0 {
		return retryDelays[0], true
	}
	if attempt > MaxRetryAttempts {
		return 0, false
	}
	idx := attempt - 1
	if idx >= len(retryDelays) {
		idx = len(retryDelays) - 1
	}
	return retryDelays[idx], true
}
