package main

import (
	_ "context"
	"fmt"
	"github.com/AleksZelenchuk/vault-server/gen/go/vaultpb"
	"github.com/AleksZelenchuk/vault-server/gen/go/vaultuserpb"
	"github.com/AleksZelenchuk/vault-server/pkg/config"
	"github.com/AleksZelenchuk/vault-server/pkg/interceptors"
	"github.com/AleksZelenchuk/vault-server/pkg/service"
	"github.com/AleksZelenchuk/vault-server/pkg/storage"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"

	_ "github.com/AleksZelenchuk/vault-server/gen/go/vaultpb"
	_ "github.com/AleksZelenchuk/vault-server/pkg/auth"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	// === Load Config ===
	_ = config.LoadConfig()
	dbURL := os.Getenv("DATABASE_URL")
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "8080"
	}
	masterKey := os.Getenv("VAULT_MASTER_KEY")
	if masterKey == "" {
		log.Fatal("VAULT_MASTER_KEY is required")
	}

	// === Connect to Database ===
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			_ = fmt.Errorf("error closing DB")
		}
	}(db)

	// === Initialize Dependencies ===
	store := storage.NewStore(db)
	userStorage := storage.NewUserStore(db)

	// === Initialize Vault Service ===
	vaultService := service.NewVaultService(store)
	userService := service.NewUserVaultService(userStorage)

	// === Set up gRPC Server with Auth Middleware ===
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.UnaryAuthInterceptor),
		grpc.ChainStreamInterceptor(interceptors.StreamAuthInterceptor),
	)
	reflection.Register(server)
	vaultuserpb.RegisterVaultUserServiceServer(server, userService)
	vaultpb.RegisterVaultServiceServer(server, vaultService)

	// === Start Listener ===
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}
	log.Printf("Vault gRPC server listening on port %s", grpcPort)

	// === Serve ===
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
