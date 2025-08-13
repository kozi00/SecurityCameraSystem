package handlers

import (
	"net/http"
	"time"
	"webserver/internal/logger"
	"webserver/internal/services"

	"github.com/gorilla/websocket"
)

// Upgrader upgrades HTTP connections to WebSocket; CheckOrigin allows all origins.
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// CameraWebsocketHandler handles camera connections over WebSocket.
// Binary messages are treated as JPEG frames and forwarded to the Manager.
// Text messages are logged for diagnostics.
func CameraWebsocketHandler(manager *services.Manager, logger *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		camera := r.URL.Query().Get("id")

		connection, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("WebSocket upgrade error: %v", err)
			return
		}
		connection.SetReadLimit(1024 * 1024)                   //Setting read limit to 1 MB
		connection.SetPingHandler(func(appData string) error { //Setting time limit for response in case connection is dead
			err := connection.SetReadDeadline(time.Now().Add(30 * time.Second))
			if err != nil {
				logger.Error("Error setting read deadline: %v", err)
			}
			return connection.WriteMessage(websocket.PongMessage, []byte(appData))
		})
		defer connection.Close()

		logger.Info("Camera connected: %s", camera)

		for {
			messageType, msg, err := connection.ReadMessage()
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

// ViewWebsocketHandler handles viewer connections over WebSocket and
// registers them in the HubService to receive broadcast frames.
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
