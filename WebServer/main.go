package main

import (
	"fmt"
	"net/http"

	"webserver/internal/handlers"
	"webserver/internal/middleware"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/view", handlers.ViewWebsocketHandler)
	mux.HandleFunc("/camera", handlers.CameraWebsocketHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// Owiń mux w AuthMiddleware
	handler := middleware.AuthMiddleware(mux)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("Access requires login with password: sienkiewicza2")
	http.ListenAndServe("0.0.0.0:8080", handler)
}
