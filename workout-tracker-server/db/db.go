package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type PostgresDb struct {
	db *pgxpool.Pool
}

func NewPostgresDb(conn string) *PostgresDb {
	dbPool, err := pgxpool.New(context.Background(), conn)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	err = dbPool.Ping(context.Background())
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	return &PostgresDb{db: dbPool}
}
