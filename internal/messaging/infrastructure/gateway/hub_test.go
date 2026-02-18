package gateway

import (
	"encoding/json"
	"testing"
	"time"

	"CONVERDA/global"
	"CONVERDA/pkg/logger"

	"CONVERDA/pkg/setting"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	global.Logger = logger.NewLogger(setting.LoggerSetting{LogLevel: "debug"})
}

func newTestHub() *Hub {
	hub := NewHub()
	go hub.Run()
	return hub
}

func newTestClient(hub *Hub, userID uuid.UUID, envID uuid.UUID) *Client {
	return &Client{
		Hub:           hub,
		send:          make(chan []byte, 256),
		UserID:        userID,
		EnvironmentID: envID,
	}
}

// wait waits a short duration for hub goroutine to process.
func wait() { time.Sleep(15 * time.Millisecond) }

func TestHub_RegisterAndBroadcastToUser(t *testing.T) {
	hub := newTestHub()
	userID := uuid.New()
	client := newTestClient(hub, userID, uuid.New())

	hub.register <- client
	wait()

	msg := []byte("hello")
	hub.BroadcastToUsers(msg, []uuid.UUID{userID})

	select {
	case received := <-client.send:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for message")
	}
}

func TestHub_UnregisterStopsDelivery(t *testing.T) {
	hub := newTestHub()
	userID := uuid.New()
	client := newTestClient(hub, userID, uuid.New())

	hub.register <- client
	wait()

	hub.unregister <- client
	wait()

	hub.BroadcastToUsers([]byte("hello"), []uuid.UUID{userID})
	select {
	case _, ok := <-client.send:
		if ok {
			t.Fatal("Should not receive message after unregister")
		}
	case <-time.After(50 * time.Millisecond):
		// Expected: no message
	}
}

func TestHub_BroadcastToEnvironment(t *testing.T) {
	hub := newTestHub()
	envID := uuid.New()

	c1 := newTestClient(hub, uuid.New(), envID)
	c2 := newTestClient(hub, uuid.New(), envID)
	c3 := newTestClient(hub, uuid.New(), uuid.New()) // different env

	hub.register <- c1
	hub.register <- c2
	hub.register <- c3
	wait()

	msg := []byte(`{"type":"test"}`)
	hub.BroadcastToEnvironment(envID, msg)

	// c1 and c2 should receive
	for _, c := range []*Client{c1, c2} {
		select {
		case received := <-c.send:
			assert.Equal(t, msg, received)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Timeout: client %s did not receive env broadcast", c.UserID)
		}
	}

	// c3 should NOT receive
	select {
	case <-c3.send:
		t.Fatal("Client in different env should not receive message")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestHub_BroadcastToEnvironmentExclude(t *testing.T) {
	hub := newTestHub()
	envID := uuid.New()

	sender := newTestClient(hub, uuid.New(), envID)
	receiver := newTestClient(hub, uuid.New(), envID)

	hub.register <- sender
	hub.register <- receiver
	wait()

	msg := []byte(`{"type":"typing_start"}`)
	hub.BroadcastToEnvironmentExclude(envID, msg, sender)

	// receiver should get it
	select {
	case received := <-receiver.send:
		assert.Equal(t, msg, received)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Receiver did not get excluded broadcast")
	}

	// sender should NOT get it
	select {
	case <-sender.send:
		t.Fatal("Sender should not receive their own excluded broadcast")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestHub_HandleInbound_TypingStart(t *testing.T) {
	hub := newTestHub()
	envID := uuid.New()

	sender := newTestClient(hub, uuid.New(), envID)
	receiver := newTestClient(hub, uuid.New(), envID)

	hub.register <- sender
	hub.register <- receiver
	wait()

	threadID := uuid.New()
	event := &WsInboundEvent{
		Type:     EventTypeTypingStart,
		ThreadID: threadID,
	}
	hub.HandleInbound(sender, event)

	// receiver should get typing event
	select {
	case received := <-receiver.send:
		var parsed map[string]interface{}
		err := json.Unmarshal(received, &parsed)
		assert.NoError(t, err)
		assert.Equal(t, EventTypeTypingStart, parsed["type"])

		payload := parsed["payload"].(map[string]interface{})
		assert.Equal(t, sender.UserID.String(), payload["user_id"])
		assert.Equal(t, threadID.String(), payload["thread_id"])

		// Should have id and timestamp
		assert.NotEmpty(t, parsed["id"])
		assert.NotEmpty(t, parsed["timestamp"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Receiver did not get typing event")
	}

	// sender should NOT get it
	select {
	case <-sender.send:
		t.Fatal("Sender should not receive their own typing event")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestHub_HandleInbound_TypingStop(t *testing.T) {
	hub := newTestHub()
	envID := uuid.New()

	sender := newTestClient(hub, uuid.New(), envID)
	receiver := newTestClient(hub, uuid.New(), envID)

	hub.register <- sender
	hub.register <- receiver
	wait()

	event := &WsInboundEvent{
		Type:     EventTypeTypingStop,
		ThreadID: uuid.New(),
	}
	hub.HandleInbound(sender, event)

	select {
	case received := <-receiver.send:
		var parsed map[string]interface{}
		err := json.Unmarshal(received, &parsed)
		assert.NoError(t, err)
		assert.Equal(t, EventTypeTypingStop, parsed["type"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Receiver did not get typing_stop event")
	}
}

func TestHub_HandleInbound_UnknownEvent(t *testing.T) {
	hub := newTestHub()
	envID := uuid.New()

	sender := newTestClient(hub, uuid.New(), envID)
	receiver := newTestClient(hub, uuid.New(), envID)

	hub.register <- sender
	hub.register <- receiver
	wait()

	event := &WsInboundEvent{
		Type: "unknown_event",
	}
	hub.HandleInbound(sender, event)

	// No message should be broadcast
	select {
	case <-receiver.send:
		t.Fatal("Should not receive anything for unknown event")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestHub_MultipleTabsSameUser(t *testing.T) {
	hub := newTestHub()
	userID := uuid.New()
	envID := uuid.New()

	tab1 := newTestClient(hub, userID, envID)
	tab2 := newTestClient(hub, userID, envID)

	hub.register <- tab1
	hub.register <- tab2
	wait()

	msg := []byte("multi-tab-msg")
	hub.BroadcastToUsers(msg, []uuid.UUID{userID})

	// Both tabs should receive
	for i, tab := range []*Client{tab1, tab2} {
		select {
		case received := <-tab.send:
			assert.Equal(t, msg, received)
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Tab %d did not receive message", i+1)
		}
	}
}
