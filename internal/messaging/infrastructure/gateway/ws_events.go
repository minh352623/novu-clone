package gateway

import (
	"encoding/json"

	"github.com/google/uuid"
)

// --- Inbound Event Types (Client → Server) ---

const (
	EventTypeTypingStart = "typing_start"
	EventTypeTypingStop  = "typing_stop"
)

// WsInboundEvent is the envelope for messages sent from client to server.
type WsInboundEvent struct {
	Type     string          `json:"type"`
	ThreadID uuid.UUID       `json:"thread_id"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// --- Outbound Event Types (Server → Client) ---
// These are already defined as string literals in broadcastEvent calls.
// Collected here for documentation.

const (
	OutboundMessageReceived  = "message_received"
	OutboundMessageReplied   = "message_replied"
	OutboundThreadAssigned   = "thread_assigned"
	OutboundThreadUnassigned = "thread_unassigned"
	OutboundThreadResolved   = "thread_resolved"
	OutboundThreadReopened   = "thread_reopened"
	OutboundThreadUpdated    = "thread_updated"
	OutboundTypingStart      = "typing_start"
	OutboundTypingStop       = "typing_stop"
)
