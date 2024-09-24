package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
	auth "proto/auth/v1/generated"
)

func main() {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		log.Fatalf("LISTEN_ADDR not set")
	}
	authSrvcAddr := os.Getenv("AUTH_SERVER_ADDR")
	if authSrvcAddr == "" {
		log.Fatalf("AUTH_SERVER_ADDR not set")
	}

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := auth.RegisterAuthorizationServiceHandlerFromEndpoint(context.Background(), mux, authSrvcAddr, opts)
	if err != nil {
		log.Fatalf("error registering auth service handler: %v", err)
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe(listenAddr, mux)
	if err != nil {
		log.Fatalf("error starting HTTP server %v", err)
	}
}
