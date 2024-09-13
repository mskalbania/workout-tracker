package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
)

// for simplicity, typically public private key pair would be used
var signSecret = []byte("secret")

type TokenAPI struct {
	registeredUsers map[string]string
}

func NewTokenAPI(registeredUsers map[string]string) *TokenAPI {
	return &TokenAPI{registeredUsers: registeredUsers}
}

type tokenRequest struct {
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type errResponse struct {
	Message string `json:"message"`
}

func (t *TokenAPI) IssueTokenHandler(c *gin.Context) {
	rq := new(tokenRequest)
	err := c.ShouldBindJSON(rq)
	if err != nil {
		abortWithError(400, err, c)
		return
	}
	pass, ok := t.registeredUsers[rq.User]
	if !ok || pass != rq.Password {
		abortWithError(401, fmt.Errorf("user unautorized"), c)
		return
	}
	claims := struct {
		Username string `json:"username"`
		jwt.StandardClaims
	}{
		Username: rq.User,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(signSecret)
	if err != nil {
		abortWithError(500, err, c)
		return
	}
	c.JSON(200, tokenResponse{Token: signedToken})
}

func abortWithError(status int, err error, c *gin.Context) {
	c.JSON(status, errResponse{Message: err.Error()})
	c.Error(err)
	c.Abort()
}
