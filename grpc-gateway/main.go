package main

import (
	"context"
	_ "github.com/envoyproxy/protoc-gen-validate/validate" //transitively required by .pb.go
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net/http"
	"os"
	auth "proto/auth/v1/generated"
	workout "proto/workout/v1/generated"
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
	workoutSrcAddr := os.Getenv("WORKOUT_SERVER_ADDR")
	if workoutSrcAddr == "" {
		log.Fatalf("WORKOUT_SERVER_ADDR not set")
	}

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(UnaryErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{EmitUnpopulated: false},
		}),
	)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := auth.RegisterAuthorizationServiceHandlerFromEndpoint(context.Background(), mux, authSrvcAddr, opts)
	if err != nil {
		log.Fatalf("error registering auth service handler: %v", err)
	}
	err = workout.RegisterExerciseServiceHandlerFromEndpoint(context.Background(), mux, workoutSrcAddr, opts)
	if err != nil {
		log.Fatalf("error registering exercise service handler: %v", err)
	}
	err = workout.RegisterWorkoutServiceHandlerFromEndpoint(context.Background(), mux, workoutSrcAddr, opts)
	if err != nil {
		log.Fatalf("error registering workout service handler: %v", err)
	}
	err = workout.RegisterWorkoutScheduleServiceHandlerFromEndpoint(context.Background(), mux, workoutSrcAddr, opts)
	if err != nil {
		log.Fatalf("error registering workout schedule service handler: %v", err)
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = http.ListenAndServe(listenAddr, mux)
	if err != nil {
		log.Fatalf("error starting HTTP server %v", err)
	}
}
