package db

import (
	"context"
	"fmt"
	"strings"
	"workout-tracker-server/model"
)

const selectFromExercise = "SELECT * FROM exercise"

type ExerciseDb interface {
	GetExercises(muscleGroup string, category string) ([]model.Exercise, error)
}

func (p *PostgresDb) GetExercises(muscleGroup string, category string) ([]model.Exercise, error) {
	query, args := getExercisesQuery(muscleGroup, category)
	rows, err := p.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exercises []model.Exercise
	for rows.Next() {
		var exercise model.Exercise
		err := rows.Scan(&exercise.ID, &exercise.Name, &exercise.Description, &exercise.Category, &exercise.MuscleGroup)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, nil
}

func getExercisesQuery(muscleGroup string, category string) (string, []any) {
	query := selectFromExercise
	var conditions []string
	var args []any
	index := 1
	if muscleGroup != "" {
		conditions = append(conditions, fmt.Sprintf("muscle_group = $%d", index))
		args = append(args, strings.ToUpper(muscleGroup))
		index++
	}
	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", index))
		args = append(args, strings.ToUpper(category))
		index++
	}
	if len(conditions) > 0 {
		query += " WHERE "
		query += strings.Join(conditions, " AND ")
	}
	return query, args
}
