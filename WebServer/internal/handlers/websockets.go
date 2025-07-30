package handlers

import (
	"log"
	"net/http"
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
		defer connection.Close()

		log.Printf("Camera connected: %s", camera)

		for {
			_, msg, err := connection.ReadMessage()
			if err != nil {
				log.Printf("Error reading camera message: %v", err)
				break
			}

			manager.HandleCameraImage(msg, camera)
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
		defer connection.Close()

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
