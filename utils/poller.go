package utils

import (
	"time"
)

// Poller runs fn every interval seconds until quit channel is closed
func Poller(interval time.Duration, quit <-chan struct{}, fn func()) {
	ticker := time.NewTicker(interval * time.Second)
	defer ticker.Stop() // Cleanup ticker when done
	
	for {
		select {
		case <-ticker.C:
			fn() // Execute the callback
		case <-quit:
			return // Stop the poller when quit signal received
		}
	}
}