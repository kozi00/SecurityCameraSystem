package websocket

import (
	"sync"
	"webserver/internal/config"
	"webserver/internal/logger"

	"github.com/gorilla/websocket"
)

type HubService struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	logger     *logger.Logger
}

func NewHubService(config *config.Config, logger *logger.Logger) *HubService {
	return &HubService{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		logger:     logger,
	}
}

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

func (h *HubService) Register(client *websocket.Conn) {
	h.register <- client
}

func (h *HubService) Unregister(client *websocket.Conn) {
	h.unregister <- client
}

func (h *HubService) Broadcast(message []byte, camera string) {
	h.broadcast <- message
}

func (h *HubService) GetClients() map[*websocket.Conn]bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients := make(map[*websocket.Conn]bool)
	for k, v := range h.clients {
		clients[k] = v
	}
	return clients
}

func (h *HubService) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}
