//go:build testing

package systems

import "time"

// Integration tests assert shortly after a websocket agent connects.
// Disable the startup jitter there so the first update runs immediately.
func getJitter() <-chan time.Time {
	ch := make(chan time.Time)
	close(ch)
	return ch
}
