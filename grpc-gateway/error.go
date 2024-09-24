package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"net/http"
)

type Error struct {
	Message string `json:"message"`
	Details []any  `json:"details,omitempty"`
}

func ErrorHandler(_ context.Context, _ *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	grpcStatus, _ := status.FromError(err)
	customError := Error{
		Message: grpcStatus.Message(),
		Details: grpcStatus.Details(),
	}

	w.Header().Set("Content-Type", m.ContentType("application/json"))
	w.WriteHeader(runtime.HTTPStatusFromCode(grpcStatus.Code()))
	err = m.NewEncoder(w).Encode(customError)
	if err != nil {
		grpclog.Errorf("faled to marshal error message: %v", err)
	}
}
