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

const ADDRESS = ":8080"

func main() {
	jwtSignKey := os.Getenv("JWT_SIGNING_KEY")
	if jwtSignKey == "" {
		log.Fatalf("JWT_SIGNING_KEY not set")
	}
	dbConnString := os.Getenv("DB_CONN_STRING")
	if dbConnString == "" {
		log.Fatalf("DB_CONN_STRING not set")
	}
	userDb := db.NewPostgresDb(dbConnString)
	userAPI := api.NewAuthorizationAPI(userDb, api.JWTProperties{
		SigningKey:           []byte(jwtSignKey),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
	})

	lis, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	s := grpc.NewServer()
	auth.RegisterAuthorizationServer(s, userAPI)

	//for debugging purposes, reflection allows (generic) clients to query for available services, types etc.
	reflection.Register(s)

	log.Println("starting server on", ADDRESS)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("error serving: %v", err)
	}
}
