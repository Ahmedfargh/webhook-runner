package main

import (
	"context"
	"log"
	"os"

	"webhookRunner/internal/app"
	"webhookRunner/internal/config"
)

func main() {
	log.Println("Starting Webhook Runner Microservice...")

	cfg := config.LoadConfig()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("Webhook Runner application bootstrap failed: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		log.Fatalf("Webhook Runner application exited with error: %v", err)
		os.Exit(1)
	}
}
