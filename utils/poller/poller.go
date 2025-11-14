package poller

import (
	"fmt"
	"time"
)

// Poller runs fn every interval until quit channel is closed
func Poller(interval time.Duration, quit <-chan struct{}, fn func()) {
	// fmt.Println("Poller started, running every", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fn()

	for {
		select {
		case <-ticker.C:
			fn()
		case <-quit:
			fmt.Println("Poller stopped via quit signal")
			return
		}
	}
}
