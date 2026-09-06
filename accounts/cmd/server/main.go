package main

import (
	"context"
	"log"
	"os"

	"accounts/internal/app"
	"accounts/internal/config"
)

func main() {
	log.Println("Starting Accounts Microservice...")

	cfg := config.LoadConfig()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Accounts application bootstrap failed: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Accounts application exited with error: %v", err)
		os.Exit(1)
	}
}
