package main

import (
	"L0WB/internal/app"
	"log"
)

func main() {
	application := app.NewApp()

	if err := application.Initialize(); err != nil {
		log.Fatalf("Failed to initialize app: %v\n", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("Failed to run app: %v\n", err)
	}

	log.Println("App finished.")
}
