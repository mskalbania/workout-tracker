package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"slices"
	"strconv"
	"strings"
	"workout-tracker-server/model"
)

var (
	ErrWorkoutNotFound         = fmt.Errorf("workout not found")
	ErrWorkoutExerciseNotFound = fmt.Errorf("workout exercise not found")

	insertWorkoutQuery         = `INSERT INTO workout (id, owner, name, comment) VALUES ($1, $2, $3, $4)`
	insertWorkoutExerciseQuery = `INSERT INTO workout_exercise (workout_exercise_id, workout_id, exercise_id, "order", repetitions, sets, weight, comment) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	selectWorkoutExercisesByWorkoutId    = `SELECT workout_exercise_id, exercise_id, "order", repetitions, sets, weight, comment FROM workout_exercise WHERE workout_id = $1`
	selectWorkoutExercisesIdsByWorkoutId = `SELECT workout_exercise_id FROM workout_exercise WHERE workout_id = $1`
	selectWorkoutByIdQuery               = `SELECT id, owner, name, comment FROM workout WHERE id = $1`
	selectWorkoutOwnerQuery              = `SELECT owner FROM workout WHERE id = $1`
	selectWorkoutsByUserIdQuery          = `SELECT id, owner, name, comment FROM workout WHERE owner = $1`

	updateWorkoutExerciseQuery = `UPDATE workout_exercise SET exercise_id = $1, "order" = $2, repetitions = $3, sets = $4, weight = $5, comment = $6 WHERE workout_exercise_id = $7`
	updateWorkoutQuery         = `UPDATE workout SET`

	deleteWorkoutQuery    = `DELETE FROM workout WHERE id = $1`
	deleteWorkoutExercise = `DELETE FROM workout_exercise WHERE workout_exercise_id = $1`
)

type WorkoutDb interface {
	SaveWorkout(workout model.Workout) (string, error)
	GetWorkouts(userId string) ([]model.Workout, error)
	GetWorkout(id string) (model.Workout, error)
	IsWorkoutOwner(workoutId, userId string) (bool, error)
	UpdateWorkout(workout model.Workout, mask *fieldmaskpb.FieldMask) error
	DeleteWorkout(id string) error
}

func (p *PostgresDb) SaveWorkout(workout model.Workout) (string, error) {
	workoutId := uuid.New().String()
	ctx := context.Background()
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite})
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
			return "", err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return workoutId, nil
}

// UpdateWorkout updates workout with given mask. If mask contains "exercises" path, it updates exercises as well.
// Update of exercises is done in PUT fashion: any existing exercise is updated, any new exercise is added, any missing exercise is deleted.
func (p *PostgresDb) UpdateWorkout(workout model.Workout, mask *fieldmaskpb.FieldMask) error {
	//wrapping with tx to make whole operation atomic
	tx, err := p.db.BeginTx(context.Background(), pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	query, args := createUpdateWorkoutQuery(workout, mask)
	if query != "" {
		tag, err := tx.Exec(context.Background(), query, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrWorkoutNotFound
		}
	}

	if slices.Contains(mask.GetPaths(), "exercises") {
		existingExercises, err := getExistingExercises(tx, workout.ID)
		if err != nil {
			return err
		}
		for _, ex := range workout.Exercises {
			if ex.WorkoutExerciseID == "" {
				if err := saveWorkoutExercise(tx, workout.ID, ex); err != nil {
					return err
				}
				continue
			}
			if slices.Contains(existingExercises, ex.WorkoutExerciseID) {
				if err := updateWorkoutExercise(tx, ex); err != nil {
					return err
				}
				existingExercises = slices.DeleteFunc(existingExercises, func(id string) bool { return id == ex.WorkoutExerciseID })
			} else {
				return ErrWorkoutExerciseNotFound
			}
		}
		for _, id := range existingExercises {
			if _, err := tx.Exec(context.Background(), deleteWorkoutExercise, id); err != nil {
				return err
			}
		}
	}
	if err = tx.Commit(context.Background()); err != nil {
		return err
	}
	return nil
}

func createUpdateWorkoutQuery(workout model.Workout, mask *fieldmaskpb.FieldMask) (string, []any) {
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

func getExistingExercises(tx pgx.Tx, workoutId string) ([]string, error) {
	rows, err := tx.Query(context.Background(), selectWorkoutExercisesIdsByWorkoutId, workoutId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var existingExercises []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		existingExercises = append(existingExercises, id)
	}
	return existingExercises, nil
}

func updateWorkoutExercise(tx pgx.Tx, ex model.WorkoutExercise) error {
	if _, err := tx.Exec(context.Background(), updateWorkoutExerciseQuery, ex.ExerciseID, ex.Order, ex.Repetitions, ex.Sets, ex.Weight, ex.Comment, ex.WorkoutExerciseID); err != nil {
		return err
	}
	return nil
}

func saveWorkoutExercise(tx pgx.Tx, workoutId string, ex model.WorkoutExercise) error {
	id := uuid.New().String()
	if _, err := tx.Exec(context.Background(), insertWorkoutExerciseQuery, id, workoutId, ex.ExerciseID, ex.Order, ex.Repetitions, ex.Sets, ex.Weight, ex.Comment); err != nil {
		return err
	}
	return nil
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
	row := p.db.QueryRow(context.Background(), selectWorkoutByIdQuery, id)
	var workout model.Workout
	err := row.Scan(&workout.ID, &workout.OwnerID, &workout.Name, &workout.Comment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Workout{}, ErrWorkoutNotFound
		}
		return model.Workout{}, err
	}
	rows, err := p.db.Query(context.Background(), selectWorkoutExercisesByWorkoutId, id)
	if err != nil {
		return model.Workout{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var ex model.WorkoutExercise
		err := rows.Scan(&ex.WorkoutExerciseID, &ex.ExerciseID, &ex.Order, &ex.Repetitions, &ex.Sets, &ex.Weight, &ex.Comment)
		if err != nil {
			return model.Workout{}, err
		}
		workout.Exercises = append(workout.Exercises, ex)
	}
	return workout, nil
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
