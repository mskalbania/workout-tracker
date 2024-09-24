package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"workout-tracker-server/model"
)

var (
	ErrIncorrectExerciseReferenced = fmt.Errorf("incorrect exercise referenced")
	insertWorkoutQuery             = `INSERT INTO workout (id, owner, name) VALUES ($1, $2, $3)`
	insertWorkoutExerciseQuery     = `INSERT INTO workout_exercise (workout_id, exercise_id, "order", repetitions, sets, weight) VALUES ($1, $2, $3, $4, $5, $6)`
)

type WorkoutDb interface {
	SaveWorkout(workout model.Workout) (string, error)
}

func (p *PostgresDb) SaveWorkout(workout model.Workout) (string, error) {
	workoutId := uuid.New().String()
	ctx := context.Background()
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, insertWorkoutQuery, workoutId, workout.OwnerID, workout.Name); err != nil {
		return "", err
	}
	for _, ex := range workout.Exercises {
		if _, err = tx.Exec(ctx, insertWorkoutExerciseQuery, workoutId, ex.ExerciseID, ex.Order, ex.Repetitions, ex.Sets, ex.Weight); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "22P02") {
				return "", ErrIncorrectExerciseReferenced
			}
			log.Printf("error inserting workout exercise: %v", err)
			return "", err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return workoutId, nil
}
