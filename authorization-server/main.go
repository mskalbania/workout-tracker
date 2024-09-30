package main

import (
	"authorization-server/api"
	"authorization-server/db"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"    //transitively required by .pb.go
	_ "google.golang.org/genproto/googleapis/api/annotations" //transitively required by .pb.go
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	auth "proto/auth/v1/generated"
	"syscall"
	"time"
)

func main() {
	appConf := loadAppConf()
	userDb := db.NewPostgresDb(appConf.dbConnString)
	userAPI := api.NewAuthorizationAPI(userDb, api.JWTProperties{
		SigningKey:           []byte(appConf.jwtSignKey),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
	}, api.UTCTimeProvider{})

	lis, err := net.Listen("tcp", appConf.listenAddr)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	defer lis.Close()

	s := grpc.NewServer()
	auth.RegisterAuthorizationServiceServer(s, userAPI)

	//for debugging purposes, reflection allows (generic) clients to query for available services, types etc.
	reflection.Register(s)

	log.Println("starting server on", appConf.listenAddr)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("error serving: %v", err)
		}
	}()

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-shutDown:
		log.Println("shutting down server")
		s.GracefulStop()
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
