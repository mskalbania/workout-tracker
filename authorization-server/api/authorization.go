package api

import (
	"authorization-server/db"
	"authorization-server/model"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
	auth.UnimplementedAuthorizationServiceServer
	userDb       db.UserDb
	properties   JWTProperties
	timeProvider TimeProvider
}

func NewAuthorizationAPI(userDb db.UserDb, properties JWTProperties, timeProvider TimeProvider) *AuthorizationAPI {
	return &AuthorizationAPI{userDb: userDb, properties: properties, timeProvider: timeProvider}
}

func (a *AuthorizationAPI) Register(ctx context.Context, rq *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	_, err := a.userDb.Find(rq.Username)
	if err == nil {
		return nil, status.Error(codes.InvalidArgument, "user already exists")
	}
	if errors.Is(err, db.ErrUserNotFound) {
		hashedPassword, err := hashPassword(rq.Password)
		if err != nil {
			log.Printf("error hashing password: %v", err)
			return nil, status.Error(codes.Internal, "error hashing password")

		}
		saved, err := a.userDb.Save(model.User{
			Username:     rq.Username,
			PasswordHash: hashedPassword,
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

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
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
	invalid := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(rq.Password))
	if invalid != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	accessToken, err := generateJWT(user.Id, a.properties, a.timeProvider)
	if err != nil {
		return nil, status.Error(codes.Internal, "error generating access token")
	}
	return &auth.LoginResponse{
		Token: accessToken,
	}, nil
}

func generateJWT(userId string, properties JWTProperties, timeProvider TimeProvider) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userId,
		IssuedAt:  jwt.NewNumericDate(timeProvider.Now()),
		ExpiresAt: jwt.NewNumericDate(timeProvider.Now().Add(properties.AccessTokenDuration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(properties.SigningKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
