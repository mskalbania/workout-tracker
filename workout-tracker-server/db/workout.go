package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	workout "proto/workout/v1/generated"
)

var (
	ErrIncorrectExerciseReferenced = fmt.Errorf("incorrect exercise referenced")
	insertWorkoutQuery             = `INSERT INTO workout (id, owner, name) VALUES ($1, $2, $3)`
	insertWorkoutExerciseQuery     = `INSERT INTO workout_exercise (workout_id, exercise_id, "order", repetitions, sets, weight) VALUES ($1, $2, $3, $4, $5, $6)`
)

type WorkoutDb interface {
	SaveWorkout(workout *workout.CreateWorkoutRequest) (string, error)
}

func (p *PostgresDb) SaveWorkout(workout *workout.CreateWorkoutRequest) (string, error) {
	workoutId := uuid.New().String()
	tx, err := p.db.BeginTx(context.Background(), pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(context.Background())
	_, err = tx.Exec(context.Background(), insertWorkoutQuery, workoutId, workout.Owner, workout.Name)
	for _, exercises := range workout.Exercises {
		_, err = tx.Exec(context.Background(),
			insertWorkoutExerciseQuery,
			workoutId,
			exercises.GetExerciseId(),
			exercises.GetOrder(),
			exercises.GetRepetitions(),
			exercises.GetSets(),
			exercises.Weight)
		if err != nil {
			var pgErr *pgconn.PgError
			//either violation of foreign key constraint or invalid uuid representation
			if errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "22P02") {
				return "", ErrIncorrectExerciseReferenced
			}
			log.Printf("error inserting workout exercise: %v", err)
			return "", err
		}
	}
	if err != nil {
		return "", err
	}
	err = tx.Commit(context.Background())
	if err != nil {
		return "", err
	}
	return workoutId, nil
}
