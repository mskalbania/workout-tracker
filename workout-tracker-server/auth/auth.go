package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"slices"
	"strings"
)

var (
	userIdCtxKey          = "userId"
	errInvalidTokenFormat = status.Errorf(codes.Unauthenticated, "invalid token - invalid format")
	errMissingToken       = status.Errorf(codes.Unauthenticated, "missing token")
	errExpiredToken       = status.Errorf(codes.Unauthenticated, "invalid token - expired")
	errMissingClaims      = status.Errorf(codes.Unauthenticated, "invalid token - missing claims")
	requiresAuth          = []string{
		"/WorkoutService/CreateWorkout",
	}
)

type Authorization struct {
	SigningKey []byte
}

func NewAuthorization(signingKey string) *Authorization {
	return &Authorization{[]byte(signingKey)}
}

func (a *Authorization) Interceptor(ctx context.Context, rq any, i *grpc.UnaryServerInfo, h grpc.UnaryHandler) (interface{}, error) {
	if slices.Contains(requiresAuth, i.FullMethod) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			log.Println("error getting metadata from context")
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}
		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, errMissingToken
		}
		userId, err := parseJWT(authHeader[0], a.SigningKey)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, userIdCtxKey, userId)
	}
	return h(ctx, rq)
}

func GetUserId(ctx context.Context) (string, error) {
	u, ok := ctx.Value(userIdCtxKey).(string)
	if !ok || u == "" {
		return "", errors.New("user id not found in context")
	}
	return u, nil
}

func parseJWT(authHeader string, key []byte) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", errInvalidTokenFormat
	}

	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return key, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", errExpiredToken
		}
		log.Println("error parsing token:", err)
		return "", errInvalidTokenFormat
	}

	subject, err := token.Claims.GetSubject()
	if err != nil || subject == "" {
		return "", errMissingClaims
	}
	return subject, nil
}
