package main

import (
	"log"
	"webserver/internal/app"
)

func main() {
	application := app.NewApp()

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
