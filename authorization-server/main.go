package main

import (
	"authorization-server/api"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	g := gin.Default()
	tokenAPI := api.NewTokenAPI(map[string]string{
		"test-user": "qwerty",
	})
	g.POST("/token", tokenAPI.IssueTokenHandler)

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: g.Handler(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}
