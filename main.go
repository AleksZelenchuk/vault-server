package main

import (
	"fmt"
	"github.com/AleksZelenchuk/vault-server/pkg/config"
	"github.com/AleksZelenchuk/vault-server/pkg/interceptors"
	"github.com/AleksZelenchuk/vault-server/pkg/storage"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	err := storage.InitCrypto()

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	_ = storage.NewStore(db)

	//srv := service.NewVaultService(store, publisher)

	/*	if err != nil {
		log.Fatal(err)
	}*/

	fmt.Printf("New server is starting")
	grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.UnaryAuthInterceptor),
		grpc.ChainStreamInterceptor(interceptors.StreamAuthInterceptor),
	)
}
