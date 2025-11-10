package utils

import "time"

// Poller runs fn every interval until quit channel is closed.
// The caller should pass an interval as a time.Duration (e.g. 1*time.Second).
func Poller(interval time.Duration, quit <-chan struct{}, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fn()
		case <-quit:
			return
		}
	}
}
