package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	auth "proto/auth/v1/generated"
)

func main() {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := auth.RegisterAuthorizationServiceHandlerFromEndpoint(context.Background(), mux, "localhost:8080", opts)
	if err != nil {
		log.Fatalf("error registering auth service handler: %v", err)
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe("localhost:8888", mux)
	if err != nil {
		log.Fatalf("error starting HTTP server")
	}
}
