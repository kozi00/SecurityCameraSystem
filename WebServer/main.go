package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // zezwala na wszystkie źródła
}
var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	camera := r.URL.Query().Get("camera")
	log.Printf("Received upload from camera: %s\n", camera)

	// Sprawdź czy ContentLength jest dostępne
	if r.ContentLength <= 0 {
		log.Printf("Invalid content length: %d", r.ContentLength)
		http.Error(w, "Invalid content length", http.StatusBadRequest)
		return
	}

	// Bezpieczne odczytanie całego body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	log.Printf("Read %d bytes from camera %s (expected %d)", len(body), camera, r.ContentLength)

	// Zamień obrazek na base64
	encoded := base64.StdEncoding.EncodeToString(body)

	// JSON do wysłania
	msg := fmt.Sprintf(`{"camera":"%s","image":"%s"}`, camera, encoded)

	mu.Lock()
	for conn := range clients {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			// Usuń klienta z błędem połączenia
			delete(clients, conn)
		}
	}
	mu.Unlock()

	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe("0.0.0.0:8080", nil)
}
