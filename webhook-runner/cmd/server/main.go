package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"webhookRunner/internal/config"
	"webhookRunner/internal/engine"
	"webhookRunner/internal/middleware"
	"webhookRunner/internal/modules/app"
	"webhookRunner/internal/modules/webhook"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("Starting Webhook Runner gRPC Microservice...")

	// 1. Initialize Database Connection & Auto-Migrations
	config.ConnectDB()

	// 2. Resolve Server Port
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}
	address := fmt.Sprintf(":%s", port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	// 3. Setup Token-based & Service-Whitelisted Authentication Interceptor
	allowedServicesRaw := os.Getenv("ALLOWED_SERVICES")
	var allowedServices []string
	if allowedServicesRaw != "" {
		allowedServices = strings.Split(allowedServicesRaw, ",")
	}
	authInterceptor := middleware.NewAuthInterceptor(os.Getenv("AUTH_TOKEN"), allowedServices)

	// 4. Create gRPC Server with Interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	// 5. Initialize Engine & HMVC Modules
	dispatcher := engine.NewDispatcher()
	appMod := app.NewAppModule(config.DB)
	webhookMod := webhook.NewWebhookModule(config.DB, appMod.Repository, dispatcher)

	// 6. Register gRPC Service Handlers
	appMod.RegisterGRPC(grpcServer)
	webhookMod.RegisterGRPC(grpcServer)

	// 7. Enable gRPC Server Reflection
	reflection.Register(grpcServer)

	// 8. Setup Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Webhook Runner gRPC server listening on %s\n", address)
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server encountered an error: %v", err)
		}
	}()

	sig := <-stopChan
	log.Printf("Received signal %v. Initiating graceful shutdown...\n", sig)

	grpcServer.GracefulStop()
	log.Println("Webhook Runner gRPC server stopped cleanly.")
}
