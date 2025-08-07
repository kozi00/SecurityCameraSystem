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

//TODO:
/*
--Dodac przycisk do usuwania zdjec
--Dodac mozliwosc filtrowania i sortowania zdjec
--Dodac ustawienia (moze)
--Error reading camera message (need fix)
--Lepsze rozwiazanie do kilku detectorow (szczegolnie w managerze do detect motion)

*/
