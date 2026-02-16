package gateway

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Map of userId to clients (one user can have multiple connections/tabs)
	userClients map[uuid.UUID]map[*Client]bool

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
					log.Printf("Failed to send to client %s, buffer full", client.UserID)
				}
			}
		}
	}
}
