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
	insertWorkoutQuery             = `INSERT INTO workout (id, owner, name, comment) VALUES ($1, $2, $3, $4)`
	insertWorkoutExerciseQuery     = `INSERT INTO workout_exercise (workout_exercise_id, workout_id, exercise_id, "order", repetitions, sets, weight, comment) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	selectWorkoutByIdQuery         = `SELECT w.id, w.owner, w.name, w.comment, we.workout_exercise_id, we.exercise_id, we."order", we.repetitions, we.sets, we.weight, we.comment FROM workout w JOIN workout_exercise we ON w.id = we.workout_id WHERE workout_id = $1;`
	updateWorkoutExerciseQuery     = `UPDATE workout_exercise SET`
	updateWorkoutQuery             = `UPDATE workout SET`
	selectWorkoutsByUserIdQuery    = `SELECT id, owner, name, comment FROM workout WHERE owner = $1`
	selectWorkoutOwnerQuery        = `SELECT owner FROM workout WHERE id = $1`
	deleteWorkoutQuery             = `DELETE FROM workout WHERE id = $1`
)

type WorkoutDb interface {
	SaveWorkout(workout model.Workout) (string, error)
	UpdateWorkoutExercise(exercise model.WorkoutExercise, mask *fieldmaskpb.FieldMask) error
	GetWorkouts(userId string) ([]model.Workout, error)
	GetWorkout(id string) (model.Workout, error)
	IsWorkoutOwner(workoutId, userId string) (bool, error)
	UpdateWorkout(workout model.Workout, mask *fieldmaskpb.FieldMask) error
	DeleteWorkout(id string) error
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
	if _, err = tx.Exec(ctx, insertWorkoutQuery, workoutId, workout.OwnerID, workout.Name, workout.Comment); err != nil {
		return "", err
	}
	for _, ex := range workout.Exercises {
		workoutExerciseId := uuid.New().String()
		if _, err = tx.Exec(ctx, insertWorkoutExerciseQuery, workoutExerciseId, workoutId, ex.ExerciseID, ex.Order, ex.Repetitions, ex.Sets, ex.Weight, ex.Comment); err != nil {
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

func (p *PostgresDb) UpdateWorkout(workout model.Workout, mask *fieldmaskpb.FieldMask) error {
	query, args := createUpdateQueryWorkout(workout, mask)
	if query == "" { // no fields to update
		return nil
	}
	tag, err := p.db.Exec(context.Background(), query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrWorkoutNotFound
	}
	return nil
}

func createUpdateQueryWorkout(workout model.Workout, mask *fieldmaskpb.FieldMask) (string, []any) {
	if mask == nil {
		return "", nil
	}
	query := updateWorkoutQuery
	var modifications []string
	var args []any
	index := 1
	for _, path := range mask.GetPaths() {
		switch path {
		case "name":
			modifications = append(modifications, fmt.Sprintf("name = $%d", index))
			args = append(args, workout.Name)
			index++
		case "comment":
			modifications = append(modifications, fmt.Sprintf("comment = $%d", index))
			args = append(args, workout.Comment)
			index++
		}
	}
	if len(modifications) == 0 {
		return "", nil
	}
	query += " " + strings.Join(modifications, ", ") + " WHERE id = $" + strconv.Itoa(index)
	args = append(args, workout.ID)
	return query, args
}

func (p *PostgresDb) IsWorkoutOwner(workoutId, userId string) (bool, error) {
	row := p.db.QueryRow(context.Background(), selectWorkoutOwnerQuery, workoutId)
	var owner string
	if err := row.Scan(&owner); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrWorkoutNotFound
		}
		return false, err
	}
	return owner == userId, nil
}

func (p *PostgresDb) GetWorkout(id string) (model.Workout, error) {
	rows, err := p.db.Query(context.Background(), selectWorkoutByIdQuery, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workout{}, ErrWorkoutNotFound
		}
		return model.Workout{}, err
	}
	defer rows.Close()
	var workout model.Workout
	for rows.Next() {
		var exercise model.WorkoutExercise
		err := rows.Scan(&workout.ID, &workout.OwnerID, &workout.Name, &workout.Comment, &exercise.WorkoutExerciseID, &exercise.ExerciseID, &exercise.Order, &exercise.Repetitions, &exercise.Sets, &exercise.Weight, &exercise.Comment)
		if err != nil {
			return model.Workout{}, err
		}
		workout.Exercises = append(workout.Exercises, exercise)
	}
	return workout, nil
}

func (p *PostgresDb) UpdateWorkoutExercise(exercise model.WorkoutExercise, mask *fieldmaskpb.FieldMask) error {
	query, args := createUpdateQuery(exercise, mask)
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

func createUpdateQuery(exercise model.WorkoutExercise, mask *fieldmaskpb.FieldMask) (string, []any) {
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
		case "comment":
			modifications = append(modifications, fmt.Sprintf("comment = $%d", index))
			args = append(args, exercise.Comment)
			index++
		}
	}
	if len(modifications) == 0 {
		return "", nil
	}
	query += " " + strings.Join(modifications, ", ") + " WHERE workout_exercise_id = $" + strconv.Itoa(index)
	args = append(args, exercise.WorkoutExerciseID)
	return query, args
}

func (p *PostgresDb) GetWorkouts(userId string) ([]model.Workout, error) {
	rows, err := p.db.Query(context.Background(), selectWorkoutsByUserIdQuery, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workouts []model.Workout
	for rows.Next() {
		var workout model.Workout
		err := rows.Scan(&workout.ID, &workout.OwnerID, &workout.Name, &workout.Comment)
		if err != nil {
			return nil, err
		}
		workouts = append(workouts, workout)
	}
	return workouts, nil
}

func (p *PostgresDb) DeleteWorkout(id string) error {
	_, err := p.db.Exec(context.Background(), deleteWorkoutQuery, id)
	if err != nil {
		return err
	}
	return nil
}
