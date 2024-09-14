package main

import (
	"authorization-server/api"
	"authorization-server/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	auth "proto/auth/v1/generated"
	"time"
)

func main() {
	key := os.Getenv("JWT_SIGNING_KEY")
	if key == "" {
		log.Fatalf("JWT_SIGNING_KEY not set")
	}

	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}

	userDb := db.NewInMemoryUserDb()
	userAPI := api.NewAuthorizationAPI(userDb, api.JWTProperties{
		SigningKey:           []byte(key),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
	})

	s := grpc.NewServer()
	auth.RegisterAuthorizationServer(s, userAPI)

	//for debugging purposes, allows clients to query for available services, types etc.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("error serving: %v", err)
	}
}
