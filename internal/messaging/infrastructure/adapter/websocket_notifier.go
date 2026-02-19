package adapter

import (
	"encoding/json"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/messaging/controller/dto"
	"CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/internal/messaging/infrastructure/gateway"

	"github.com/google/uuid"
)

type WebSocketNotifier struct {
	hub *gateway.Hub
}

func NewWebSocketNotifier(hub *gateway.Hub) repository.IMessageNotifier {
	return &WebSocketNotifier{hub: hub}
}

func (n *WebSocketNotifier) BroadcastEvent(envID uuid.UUID, eventType string, payload interface{}) {
	if n.hub == nil {
		return
	}

	event := dto.WsEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		global.Logger.Error("messaging_notifier: failed to marshal broadcast event", "type", eventType, "error", err)
		return
	}

	n.hub.BroadcastToEnvironment(envID, data)
}
