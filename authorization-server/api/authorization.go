package api

import (
	"authorization-server/db"
	"authorization-server/model"
	"authorization-server/server"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type JWTProperties struct {
	SigningKey           []byte
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type AuthorizationAPI struct {
	server.UnimplementedAuthorizationServer
	userDb     db.UserDb
	properties JWTProperties
}

func NewAuthorizationAPI(userDb db.UserDb, properties JWTProperties) *AuthorizationAPI {
	return &AuthorizationAPI{userDb: userDb, properties: properties}
}

func (a *AuthorizationAPI) Register(ctx context.Context, rq *server.RegisterRequest) (*server.RegisterResponse, error) {
	user, err := a.userDb.Find(rq.Username)
	if err != nil {
		log.Printf("error finding user: %v", err)
		return nil, status.Error(codes.Internal, "error finding user")
	}
	if user != nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	saved, err := a.userDb.Save(&model.User{
		Username: rq.Username,
		Password: rq.Password,
	})
	if err != nil {
		log.Printf("error saving user: %v", err)
		return nil, status.Error(codes.Internal, "error saving user")
	}
	return &server.RegisterResponse{
		UserId: saved.Id,
	}, nil
}

func (a *AuthorizationAPI) Login(ctx context.Context, rq *server.LoginRequest) (*server.LoginResponse, error) {
	user, err := a.userDb.Find(rq.Username)
	if err != nil {
		log.Printf("error finding user: %v", err)
		return nil, status.Error(codes.Internal, "error finding user")
	}
	if user == nil || user.Password != rq.Password {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	accessToken, err := generateJWT(user.Id, a.properties)
	if err != nil {
		return nil, status.Error(codes.Internal, "error generating access token")
	}
	return &server.LoginResponse{
		Token: accessToken,
	}, nil
}

func generateJWT(userId string, properties JWTProperties) (string, error) {
	claims := struct {
		UserId string `json:"user_id"`
		jwt.RegisteredClaims
	}{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(properties.AccessTokenDuration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(properties.SigningKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
