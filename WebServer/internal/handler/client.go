package handler

import (
	"net/http"
	"webserver/internal/logger"
	"webserver/internal/service"

	"github.com/gorilla/websocket"
)

// Upgrader upgrades HTTP connections to WebSocket; CheckOrigin allows all origins.
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ViewWebsocketHandler handles viewer connections over WebSocket and
// registers them in the HubService to receive broadcast frames.
func ViewWebsocketHandler(manager *service.Manager, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("WebSocket upgrade error: %v", err)
			return
		}

		manager.GetWebsocketService().Register(connection)
		defer manager.GetWebsocketService().Unregister(connection)

		logger.Info("Viewer connected")

		for {
			_, _, err := connection.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Info("Viewer disconnected normally")
				} else {
					logger.Error("Viewer disconnected with error: %v", err)
				}
				break
			}
		}
	}
}
