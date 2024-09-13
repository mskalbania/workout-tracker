package api

import (
	"authorization-server/server"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// for simplicity, typically public private key pair would be used
var signSecret = []byte("secret")

type TokenAPI struct {
	server.UnimplementedAuthorizationServer
	registeredUsers map[string]string
}

func NewTokenAPI(registeredUsers map[string]string) *TokenAPI {
	return &TokenAPI{registeredUsers: registeredUsers}
}

func (t *TokenAPI) IssueToken(ctx context.Context, rq *server.IssueTokenRequest) (*server.IssueTokenResponse, error) {
	pass, ok := t.registeredUsers[rq.Username]
	if !ok || pass != rq.Password {
		err := status.Error(codes.Unauthenticated, "invalid username or password")
		return nil, err
	}
	claims := struct {
		Username string `json:"username"`
		jwt.RegisteredClaims
	}{
		Username: rq.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(signSecret)
	if err != nil {
		return nil, err
	}
	return &server.IssueTokenResponse{Token: signedToken}, nil
}
