package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"strconv"
	"strings"
	"workout-tracker-server/model"
)

var (
	ErrIncorrectExerciseReferenced = fmt.Errorf("incorrect exercise referenced")
	ErrWorkoutNotFound             = fmt.Errorf("workout not found")
	ErrorExerciseNotFound          = fmt.Errorf("exercise not found")
	insertWorkoutQuery             = `INSERT INTO workout (id, owner, name) VALUES ($1, $2, $3)`
	insertWorkoutExerciseQuery     = `INSERT INTO workout_exercise (workout_id, exercise_id, "order", repetitions, sets, weight) VALUES ($1, $2, $3, $4, $5, $6)`
	selectWorkoutQuery             = `SELECT id, owner, name FROM workout WHERE id = $1`
	updateWorkoutExerciseQuery     = `UPDATE workout_exercise SET`
)

type WorkoutDb interface {
	SaveWorkout(workout model.Workout) (string, error)
	GetWorkout(id string) (model.Workout, error)
	UpdateWorkoutExercise(workoutId string, exercise model.WorkoutExercise, mask *fieldmaskpb.FieldMask) error
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
			if errors.As(err, &pgErr) && (pgErr.Code == "23503" || pgErr.Code == "22P02") { // foreign key violation or invalid input syntax
				return "", ErrIncorrectExerciseReferenced
			}
			return "", err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return workoutId, nil
}

func (p *PostgresDb) GetWorkout(id string) (model.Workout, error) {
	row := p.db.QueryRow(context.Background(), selectWorkoutQuery, id)
	var workout model.Workout
	if err := row.Scan(&workout.ID, &workout.OwnerID, &workout.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workout{}, ErrWorkoutNotFound
		}
		return model.Workout{}, err
	}
	return workout, nil
}

func (p *PostgresDb) UpdateWorkoutExercise(workoutId string, exercise model.WorkoutExercise, mask *fieldmaskpb.FieldMask) error {
	query, args := createUpdateQuery(workoutId, exercise, mask)
	if query == "" { // no fields to update
		return nil
	}
	tag, err := p.db.Exec(context.Background(), query, args...)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "22P02" {
		return ErrorExerciseNotFound
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrorExerciseNotFound
	}
	return nil
}

func createUpdateQuery(workoutId string, exercise model.WorkoutExercise, mask *fieldmaskpb.FieldMask) (string, []any) {
	if mask == nil {
		return "", nil
	}
	query := updateWorkoutExerciseQuery
	var modifications []string
	var args []any
	index := 1
	for _, path := range mask.GetPaths() {
		switch path {
		case "weight":
			modifications = append(modifications, fmt.Sprintf("weight = $%d", index))
			args = append(args, exercise.Weight)
			index++
		case "sets":
			modifications = append(modifications, fmt.Sprintf("sets = $%d", index))
			args = append(args, exercise.Sets)
			index++
		case "repetitions":
			modifications = append(modifications, fmt.Sprintf("repetitions = $%d", index))
			args = append(args, exercise.Repetitions)
			index++
		case "order":
			modifications = append(modifications, fmt.Sprintf("\"order\" = $%d", index))
			args = append(args, exercise.Order)
			index++
		}
	}
	if len(modifications) == 0 {
		return "", nil
	}
	query += " " + strings.Join(modifications, ", ") + " WHERE workout_id = $" + strconv.Itoa(index) + " AND exercise_id = $" + strconv.Itoa(index+1)
	args = append(args, workoutId, exercise.ExerciseID)
	return query, args
}
