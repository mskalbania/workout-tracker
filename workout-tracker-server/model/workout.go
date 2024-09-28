package model

import (
	workout "proto/workout/v1/generated"
)

type Workout struct {
	ID        string
	OwnerID   string
	Name      string
	Comment   *string
	Exercises []WorkoutExercise
}

type WorkoutExercise struct {
	WorkoutExerciseID string
	ExerciseID        string
	Order             int32
	Repetitions       int32
	Sets              int32
	Weight            *int32
	Comment           *string
}

func FromWorkoutExerciseProto(proto *workout.WorkoutExercise) WorkoutExercise {
	return WorkoutExercise{
		WorkoutExerciseID: proto.WorkoutExerciseId,
		ExerciseID:        proto.ExerciseId,
		Order:             proto.Order,
		Repetitions:       proto.Repetitions,
		Sets:              proto.Sets,
		Weight:            proto.Weight,
		Comment:           proto.Comment,
	}
}

func FromWorkoutProto(proto *workout.Workout) Workout {
	var exercises []WorkoutExercise
	for _, ex := range proto.Exercises {
		exercises = append(exercises, FromWorkoutExerciseProto(ex))
	}
	return Workout{
		ID:        proto.Id,
		Name:      proto.Name,
		Comment:   proto.Comment,
		Exercises: exercises,
	}
}

func (w Workout) ToProto() *workout.Workout {
	var exercises []*workout.WorkoutExercise
	for _, ex := range w.Exercises {
		exercises = append(exercises, ex.toProto())
	}
	return &workout.Workout{
		Id:        w.ID,
		Name:      w.Name,
		Comment:   w.Comment,
		Exercises: exercises,
	}
}

func (w WorkoutExercise) toProto() *workout.WorkoutExercise {
	return &workout.WorkoutExercise{
		WorkoutExerciseId: w.WorkoutExerciseID,
		ExerciseId:        w.ExerciseID,
		Order:             w.Order,
		Repetitions:       w.Repetitions,
		Sets:              w.Sets,
		Weight:            w.Weight,
		Comment:           w.Comment,
	}
}
