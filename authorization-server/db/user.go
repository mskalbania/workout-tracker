package db

import (
	"authorization-server/model"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

var (
	insertUser      = `INSERT INTO "user" (id, email, password_hash) VALUES ($1, $2, $3)`
	findUserByEmail = `SELECT id, email, password_hash FROM "user" WHERE email = $1`
)

var ErrUserNotFound = fmt.Errorf("user not found")

type UserDb interface {
	Save(user model.User) (model.User, error)
	Find(username string) (model.User, error)
}

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

func (i *PostgresDb) Save(user model.User) (model.User, error) {
	user.ID = uuid.New().String()
	_, err := i.db.Exec(context.Background(), insertUser, user.ID, user.Username, user.PasswordHash)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (i *PostgresDb) Find(username string) (model.User, error) {
	var user model.User
	err := i.db.QueryRow(context.Background(), findUserByEmail, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, ErrUserNotFound
	}
	return user, err
}
