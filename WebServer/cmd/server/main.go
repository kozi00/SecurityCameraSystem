package main

import (
	"log"
	"webserver/internal/app"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	application := app.NewApp()

	if err := application.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
