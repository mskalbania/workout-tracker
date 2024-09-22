package db

import (
	"context"
	"fmt"
	exercise "proto/workout/v1/generated"
	"strings"
)

const selectFromExercise = "SELECT * FROM exercise"

type ExerciseDb interface {
	GetExercises(muscleGroup string, category string) ([]*exercise.Exercise, error)
}

func (p *PostgresDb) GetExercises(muscleGroup string, category string) ([]*exercise.Exercise, error) {
	query, args := getExercisesQuery(muscleGroup, category)
	rows, err := p.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var exercises []*exercise.Exercise
	for rows.Next() {
		var ex exercise.Exercise
		err := rows.Scan(&ex.Id, &ex.Name, &ex.Description, &ex.Category, &ex.MuscleGroup)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, &ex)
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
