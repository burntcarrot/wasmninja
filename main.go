package main

import (
	"log"

	"github.com/burntcarrot/wasmninja/internal/server"
)

func main() {
	// Create the app
	app, err := server.NewApp()
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	// Start the app
	if err := app.Start(); err != nil {
		log.Fatal("Failed to start the app:", err)
	}
}
