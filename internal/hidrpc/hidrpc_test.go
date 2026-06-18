package hidrpc

import "testing"

// Keepalives must be processed on the same queue as key press/release events.
// If they land on a different queue, the per-queue consumer goroutines run
// concurrently and a keepalive can re-extend the auto-release timer after a
// key-up has already been dequeued, causing repeated keys on unstable networks.
func TestKeepAliveSharesKeyEventQueue(t *testing.T) {
	keyQueue := GetQueueIndex(TypeKeypressReport)
	keepAliveQueue := GetQueueIndex(TypeKeypressKeepAliveReport)

	if keepAliveQueue != keyQueue {
		t.Fatalf("keepalive queue (%d) must match keypress queue (%d) to preserve ordering",
			keepAliveQueue, keyQueue)
	}

	if GetQueueIndex(TypeKeyboardReport) != keyQueue {
		t.Fatalf("keyboard report queue must match keypress queue (%d)", keyQueue)
	}
}
