package main

import (
	"authorization-server/api"
	"authorization-server/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	auth "proto/auth/v1/generated"
	"time"
)

func main() {
	appConf := loadAppConf()
	userDb := db.NewPostgresDb(appConf.dbConnString)
	userAPI := api.NewAuthorizationAPI(userDb, api.JWTProperties{
		SigningKey:           []byte(appConf.jwtSignKey),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
	})

	lis, err := net.Listen("tcp", appConf.listenAddr)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	s := grpc.NewServer()
	auth.RegisterAuthorizationServer(s, userAPI)

	//for debugging purposes, reflection allows (generic) clients to query for available services, types etc.
	reflection.Register(s)

	log.Println("starting server on", appConf.listenAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("error serving: %v", err)
	}
}

type appConf struct {
	jwtSignKey   string
	dbConnString string
	listenAddr   string
}

func loadAppConf() appConf {
	return appConf{
		jwtSignKey:   loadPropertyOrFail("JWT_SIGNING_KEY"),
		dbConnString: loadPropertyOrFail("DB_CONN_STRING"),
		listenAddr:   loadPropertyOrFail("LISTEN_ADDR"),
	}
}

func loadPropertyOrFail(propName string) string {
	propVal := os.Getenv(propName)
	if propVal == "" {
		log.Fatalf("%s not set", propName)
	}
	return propVal
}
