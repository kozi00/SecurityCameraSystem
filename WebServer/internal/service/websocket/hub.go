package websocket

import (
	"sync"
	"webserver/internal/config"
	"webserver/internal/logger"

	"github.com/gorilla/websocket"
)

// HubService manages WebSocket clients and broadcasting messages to them.
type HubService struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	logger     *logger.Logger
}

// NewHubService constructs a HubService with internal channels and maps.
func NewHubService(config *config.Config, logger *logger.Logger) *HubService {
	return &HubService{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		logger:     logger,
	}
}

// Run processes register/unregister requests and broadcasts messages to all clients.
func (h *HubService) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			h.logger.Info("Client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
			h.logger.Info("Client disconnected. Total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					h.logger.Error("Error sending message: %v", err)
					delete(h.clients, client)
					client.Close()
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Register enqueues a WebSocket client for registration.
func (h *HubService) Register(client *websocket.Conn) {
	h.register <- client
}

// Unregister enqueues a WebSocket client for removal and closes it.
func (h *HubService) Unregister(client *websocket.Conn) {
	h.unregister <- client
}

// Broadcast enqueues a message for broadcasting to all clients.
func (h *HubService) Broadcast(message []byte, camera string) {
	h.broadcast <- message
}

// GetClients returns a snapshot copy of current clients.
func (h *HubService) GetClients() map[*websocket.Conn]bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make(map[*websocket.Conn]bool)
	for k, v := range h.clients {
		clients[k] = v
	}
	return clients
}

// GetClientCount returns the number of currently connected clients.
func (h *HubService) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}
