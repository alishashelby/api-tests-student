package utils

import (
	"testing"
	"time"
)

// WaitUntil повторяет проверку до успеха или истечения таймаута
func WaitUntil(t *testing.T, timeout, interval time.Duration, failureMessage string, check func() bool) {
	t.Helper()

	if check() {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			if check() {
				return
			}
		case <-timer.C:
			t.Fatalf("%s за %s", failureMessage, timeout)
		}
	}
}
