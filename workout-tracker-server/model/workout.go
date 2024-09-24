package model

import workout "proto/workout/v1/generated"

type Workout struct {
	ID        string
	OwnerID   string
	Name      string
	Exercises []WorkoutExercise
}

type WorkoutExercise struct {
	ExerciseID  string
	Order       int32
	Repetitions int32
	Sets        int32
	Weight      *int32
}

func FromProto(proto *workout.WorkoutExercise) WorkoutExercise {
	return WorkoutExercise{
		ExerciseID:  proto.ExerciseId,
		Order:       proto.Order,
		Repetitions: proto.Repetitions,
		Sets:        proto.Sets,
		Weight:      proto.Weight,
	}
}
