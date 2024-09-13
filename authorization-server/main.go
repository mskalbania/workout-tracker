package main

import (
	"authorization-server/api"
	"authorization-server/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	s := grpc.NewServer()
	server.RegisterAuthorizationServer(s, api.NewTokenAPI(map[string]string{
		"admin": "admin",
	}))

	//for debugging purposes, allows clients to query for available services, types etc.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("error serving: %v", err)
	}
}
