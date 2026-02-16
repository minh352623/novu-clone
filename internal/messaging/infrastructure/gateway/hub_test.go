package gateway

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHub_Existed(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	userID := uuid.New()
	client := &Client{
		Hub:    hub,
		send:   make(chan []byte, 256),
		UserID: userID,
	}

	// 1. Register
	hub.register <- client
	time.Sleep(10 * time.Millisecond) // Wait for registration

	// Check if registered
	// Since clients map is private, we can't check directly easily without exposing,
	// but we can check via behavior (broadcast)
	// Or we can use reflection/unsafe, but let's stick to behavior.

	// 2. Broadcast to User
	msg := []byte("hello")
	hub.BroadcastToUsers(msg, []uuid.UUID{userID})

	// 3. Receive
	select {
	case received := <-client.send:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}

	// 4. Unregister
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	// 5. Broadcast again (should not receive)
	// We might panic if hub tries to send to closed channel, but we removed it so it shouldn't.
	// But reading from closed channel returns immediately.
	hub.BroadcastToUsers(msg, []uuid.UUID{userID})
	select {
	case _, ok := <-client.send:
		if ok {
			t.Fatal("Should not receive message (open channel) after unregister")
		}
		// If !ok, channel is closed, which implies unregister happened. Success.
	case <-time.After(50 * time.Millisecond):
		// Expected timeout if channel was NOT closed and NO message sent
		// But since Unregister closes channel, we expect the first case to hit with !ok
	}
}
