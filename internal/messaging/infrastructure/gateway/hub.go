package gateway

import (
	"encoding/json"
	"sync"
	"time"

	"CONVERDA/global"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Map of userId to clients (one user can have multiple connections/tabs)
	userClients map[uuid.UUID]map[*Client]bool

	// Map of environmentId to clients
	envClients map[uuid.UUID]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Lock for map access
	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		userClients: make(map[uuid.UUID]map[*Client]bool),
		envClients:  make(map[uuid.UUID]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.userClients[client.UserID]; !ok {
				h.userClients[client.UserID] = make(map[*Client]bool)
			}
			h.userClients[client.UserID][client] = true

			if _, ok := h.envClients[client.EnvironmentID]; !ok {
				h.envClients[client.EnvironmentID] = make(map[*Client]bool)
			}
			h.envClients[client.EnvironmentID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove from userClients
				if userConns, ok := h.userClients[client.UserID]; ok {
					delete(userConns, client)
					if len(userConns) == 0 {
						delete(h.userClients, client.UserID)
					}
				}

				// Remove from envClients
				if envConns, ok := h.envClients[client.EnvironmentID]; ok {
					delete(envConns, client)
					if len(envConns) == 0 {
						delete(h.envClients, client.EnvironmentID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// Broadcast to everyone (maybe not used much but good for testing)
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToUsers sends a message to specific users
func (h *Hub) BroadcastToUsers(message []byte, userIDs []uuid.UUID) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range userIDs {
		if clients, ok := h.userClients[userID]; ok {
			for client := range clients {
				select {
				case client.send <- message:
				default:
					// If channel is full, we might want to log or disconnect
					global.Logger.Warn("ws: buffer full for client " + client.UserID.String())
				}
			}
		}
	}
}

// BroadcastToEnvironment sends a message to all clients in a specific environment
func (h *Hub) BroadcastToEnvironment(envID uuid.UUID, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.envClients[envID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				global.Logger.Warn("ws: buffer full for client " + client.UserID.String() + " in env " + envID.String())
			}
		}
	}
}

// BroadcastToEnvironmentExclude sends a message to all env clients except the excluded one.
func (h *Hub) BroadcastToEnvironmentExclude(envID uuid.UUID, message []byte, exclude *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.envClients[envID]; ok {
		for client := range clients {
			if client == exclude {
				continue
			}
			select {
			case client.send <- message:
			default:
				global.Logger.Warn("ws: buffer full for client " + client.UserID.String() + " in env " + envID.String())
			}
		}
	}
}

// HandleInbound processes a client-to-server event.
func (h *Hub) HandleInbound(sender *Client, event *WsInboundEvent) {
	switch event.Type {
	case EventTypeTypingStart, EventTypeTypingStop:
		h.handleTyping(sender, event)
	default:
		global.Logger.Warn("ws: unknown inbound event type " + string(event.Type) + " from " + sender.UserID.String())
	}
}

// handleTyping broadcasts typing indicator events to other clients in the same env.
func (h *Hub) handleTyping(sender *Client, event *WsInboundEvent) {
	outbound := map[string]interface{}{
		"id":        uuid.New().String(),
		"type":      event.Type,
		"timestamp": time.Now(),
		"payload": map[string]interface{}{
			"user_id":   sender.UserID,
			"thread_id": event.ThreadID,
		},
	}

	data, err := json.Marshal(outbound)
	if err != nil {
		return
	}

	h.BroadcastToEnvironmentExclude(sender.EnvironmentID, data, sender)
}
