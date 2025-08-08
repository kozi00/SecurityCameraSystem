package handlers

import (
	"net/http"
	"time"
	"webserver/internal/logger"
	"webserver/internal/services"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler dla kamer z zależnościami
func CameraWebsocketHandler(manager *services.Manager, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		camera := r.URL.Query().Get("id")

		connection, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("WebSocket upgrade error: %v", err)
			return
		}

		defer connection.Close()

		logger.Info("Camera connected: %s", camera)

		for {
			messageType, msg, err := connection.ReadMessage()
			connection.SetReadDeadline(time.Now().Add(30 * time.Second))

			if err != nil {
				logger.Error("Error reading camera message: %v", err)
				break
			}

			switch messageType {
			case websocket.TextMessage:
				logger.Info("Camera %s sent text message: %s", camera, msg)
			case websocket.BinaryMessage:
				manager.HandleCameraImage(msg, camera)
			default:
				logger.Warning("Camera %s sent unsupported message type: %d", camera, messageType)
			}
		}
	}
}

// Handler dla viewerów z zależnościami
func ViewWebsocketHandler(manager *services.Manager, logger *logger.Logger) http.HandlerFunc {
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
				logger.Error("Viewer disconnected: %v", err)
				break
			}
		}
	}
}
