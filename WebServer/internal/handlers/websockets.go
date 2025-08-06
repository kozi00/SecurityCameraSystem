package handlers

import (
	"log"
	"net/http"
	"time"
	"webserver/internal/services"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler dla kamer z zależnościami
func CameraWebsocketHandler(manager *services.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		camera := r.URL.Query().Get("id")

		connection, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		connection.SetPingHandler(func(appData string) error {
			//	log.Printf("Camera %s ping received", camera)
			connection.SetReadDeadline(time.Now().Add(15 * time.Second))
			return connection.WriteMessage(websocket.PongMessage, []byte(appData))
		})
		defer connection.Close()

		log.Printf("Camera connected: %s", camera)

		for {
			messageType, msg, err := connection.ReadMessage()
			if err != nil {
				log.Printf("Error reading camera message: %v", err)
				break
			}

			switch messageType {
			case websocket.TextMessage:
				log.Printf("Camera %s sent text message: %s", camera, msg)
			case websocket.BinaryMessage:
				manager.HandleCameraImage(msg, camera)
			default:
				log.Printf("Camera %s sent unsupported message type: %d", camera, messageType)
			}
		}
	}
}

// Handler dla viewerów z zależnościami
func ViewWebsocketHandler(manager *services.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, err := Upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		manager.GetWebsocketService().Register(connection)
		defer manager.GetWebsocketService().Unregister(connection)

		log.Printf("Viewer connected")

		for {
			_, _, err := connection.ReadMessage()
			if err != nil {
				log.Printf("Viewer disconnected: %v", err)
				break
			}
		}
	}
}
