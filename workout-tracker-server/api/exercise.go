package api

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	workout "proto/workout/v1/generated"
	"workout-tracker-server/db"
)

type ExerciseAPI struct {
	workout.UnimplementedExerciseServiceServer
	db db.ExerciseDb
}

func NewExerciseAPI(db db.ExerciseDb) *ExerciseAPI {
	return &ExerciseAPI{db: db}
}

func (e *ExerciseAPI) GetExercises(ctx context.Context, rq *workout.GetExercisesRequest) (*workout.GetExercisesResponse, error) {
	exercises, err := e.db.GetExercises(rq.GetMuscleGroupFilter(), rq.GetCategoryFilter())
	if err != nil {
		log.Printf("error getting exercises: %v", err)
		return nil, status.Error(codes.Internal, "error getting exercises")
	}
	return &workout.GetExercisesResponse{Exercises: exercises}, nil
}
