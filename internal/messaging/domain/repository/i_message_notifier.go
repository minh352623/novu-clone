package repository

import (
	"github.com/google/uuid"
)

// IMessageNotifier defines a port for broadcasting real-time messaging events.
type IMessageNotifier interface {
	BroadcastEvent(envID uuid.UUID, eventType string, payload interface{})
}
