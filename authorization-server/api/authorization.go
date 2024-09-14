package api

import (
	"authorization-server/db"
	"authorization-server/model"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	auth "proto/auth/v1/generated"
	"time"
)

type JWTProperties struct {
	SigningKey           []byte
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type AuthorizationAPI struct {
	auth.UnimplementedAuthorizationServer
	userDb     db.UserDb
	properties JWTProperties
}

func NewAuthorizationAPI(userDb db.UserDb, properties JWTProperties) *AuthorizationAPI {
	return &AuthorizationAPI{userDb: userDb, properties: properties}
}

func (a *AuthorizationAPI) Register(ctx context.Context, rq *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	_, err := a.userDb.Find(rq.Username)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	if errors.Is(err, db.ErrUserNotFound) {
		saved, err := a.userDb.Save(model.User{
			Username: rq.Username,
			Password: rq.Password,
		})
		if err != nil {
			log.Printf("error saving user: %v", err)
			return nil, status.Error(codes.Internal, "error saving user")
		}
		return &auth.RegisterResponse{
			UserId: saved.Id,
		}, nil
	}
	log.Printf("error finding user: %v", err)
	return nil, status.Error(codes.Internal, "error finding user")
}

func (a *AuthorizationAPI) Login(ctx context.Context, rq *auth.LoginRequest) (*auth.LoginResponse, error) {
	user, err := a.userDb.Find(rq.Username)
	if errors.Is(err, db.ErrUserNotFound) {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}
	if err != nil {
		log.Printf("error finding user: %v", err)
		return nil, status.Error(codes.Internal, "error finding user")
	}
	if user.Password != rq.Password {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	accessToken, err := generateJWT(user.Id, a.properties)
	if err != nil {
		return nil, status.Error(codes.Internal, "error generating access token")
	}
	return &auth.LoginResponse{
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
