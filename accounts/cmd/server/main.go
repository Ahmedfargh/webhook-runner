package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"accounts/internal/config"
	"accounts/internal/middleware"
	"accounts/internal/modules/admin"
	"accounts/internal/modules/permission"
	"accounts/internal/modules/role"
	"accounts/internal/modules/user"
	"accounts/internal/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("Starting Accounts gRPC Service...")

	// 1. Initialize Database Connection
	config.ConnectDB()

	// 2. Resolve Server Port
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	address := fmt.Sprintf(":%s", port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	// 3. Setup Token-based Authentication Interceptor (Bearer token)
	authInterceptor := middleware.NewAuthInterceptor(os.Getenv("AUTH_TOKEN"))

	// 4. Create gRPC Server with Interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	// 4. Initialize Shared Repositories & HMVC Modules
	countryRepo := repository.NewCountryRepository(config.DB)

	permissionMod := permission.NewPermissionModule(config.DB)
	roleMod := role.NewRoleModule(config.DB, permissionMod.Repository, permissionMod.Presenter)
	userMod := user.NewUserModule(config.DB, countryRepo)
	adminMod := admin.NewAdminModule(
		config.DB,
		countryRepo,
		roleMod.Repository,
		permissionMod.Repository,
		roleMod.Presenter,
		permissionMod.Presenter,
	)

	// 5. Register gRPC Service Handlers
	permissionMod.RegisterGRPC(grpcServer)
	roleMod.RegisterGRPC(grpcServer)
	userMod.RegisterGRPC(grpcServer)
	adminMod.RegisterGRPC(grpcServer)

	// 6. Enable gRPC Server Reflection (allows Postman & grpcurl inspection)
	reflection.Register(grpcServer)

	// 7. Setup Graceful Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Accounts gRPC server listening on %s\n", address)
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("gRPC server encountered an error: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-stopChan
	log.Printf("Received signal %v. Initiating graceful shutdown...\n", sig)

	grpcServer.GracefulStop()
	log.Println("gRPC server stopped cleanly.")
}
