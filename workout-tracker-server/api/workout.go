package api

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	workout "proto/workout/v1/generated"
	"workout-tracker-server/auth"
	"workout-tracker-server/db"
	"workout-tracker-server/model"
)

type WorkoutAPI struct {
	workout.UnimplementedWorkoutServiceServer
	db db.WorkoutDb
}

func NewWorkoutAPI(db db.WorkoutDb) *WorkoutAPI {
	return &WorkoutAPI{db: db}
}

func (w *WorkoutAPI) CreateWorkout(ctx context.Context, rq *workout.CreateWorkoutRequest) (*workout.CreateWorkoutResponse, error) {
	userId, err := auth.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	if userId != rq.GetOwner() {
		return nil, status.Error(codes.PermissionDenied, "access forbidden")
	}
	id, err := w.db.SaveWorkout(toWorkout(rq))
	if err != nil {
		if errors.Is(err, db.ErrIncorrectExerciseReferenced) {
			return nil, status.Error(codes.InvalidArgument, "incorrect exercise referenced")
		}
		log.Printf("error saving workout: %v", err)
		return nil, status.Error(codes.Internal, "error saving workout")
	}
	return &workout.CreateWorkoutResponse{
		Id: id,
	}, nil
}

func toWorkout(rq *workout.CreateWorkoutRequest) model.Workout {
	var exercises []model.WorkoutExercise
	for _, ex := range rq.Exercises {
		exercises = append(exercises, model.FromProto(ex))
	}
	return model.Workout{
		Owner:     rq.Owner,
		Name:      rq.Name,
		Exercises: exercises,
	}
}
