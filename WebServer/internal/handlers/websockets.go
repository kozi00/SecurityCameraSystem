package handlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sync"
	"webserver/internal/services"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var Viewers = make(map[*websocket.Conn]bool)
var Mu sync.Mutex
var detectionService = services.NewObjectDetectionService()

func CameraWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	camera := r.URL.Query().Get("id")

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			Mu.Lock()
			delete(Viewers, conn)
			Mu.Unlock()
			break
		}
		SendImageFromCameraToClients(camera, msg)
	}
}

func ViewWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	Mu.Lock()
	Viewers[conn] = true
	Mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			Mu.Lock()
			delete(Viewers, conn)
			Mu.Unlock()
			break
		}
	}
}

func SendImageFromCameraToClients(camera string, image []byte) {
	motionDetected, err := detectionService.DetectMotion(image)
	if err != nil {
		log.Printf("Błąd rozpoznawania obiektów: %v", err)
	}

	var detectionsJSON string

	if motionDetected {
		detections, _ := detectionService.DetectObjects(image)

		detectionsJSON, err = detectionService.FormatDetectionsAsJSON(detections)
		if err != nil {
			log.Printf("Błąd formatowania detekcji do JSON: %v", err)
		}
	} else {
		detectionsJSON = "[]"
	}

	encoded := base64.StdEncoding.EncodeToString(image)

	msg := fmt.Sprintf(`{"camera":"%s","image":"%s", "detections":%s}`, camera, encoded, detectionsJSON)

	Mu.Lock()
	defer Mu.Unlock()

	for conn := range Viewers {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			delete(Viewers, conn)
		}
	}
}
