package api

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
	id, err := w.db.SaveWorkout(toWorkout(userId, rq))
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

func toWorkout(userId string, rq *workout.CreateWorkoutRequest) model.Workout {
	var exercises []model.WorkoutExercise
	for _, ex := range rq.Exercises {
		exercises = append(exercises, model.FromProto(ex))
	}
	return model.Workout{
		OwnerID:   userId,
		Name:      rq.Name,
		Exercises: exercises,
	}
}

func (w *WorkoutAPI) UpdateWorkoutExercise(ctx context.Context, rq *workout.UpdateWorkoutRequest) (*emptypb.Empty, error) {
	uuid, err := auth.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	wrk, err := w.db.GetWorkout(rq.Exercise.WorkoutId)
	if errors.Is(err, db.ErrWorkoutNotFound) {
		return nil, status.Error(codes.NotFound, "workout not found")
	}
	if err != nil {
		log.Printf("error getting workout: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout")
	}
	if wrk.OwnerID != uuid {
		return nil, status.Error(codes.PermissionDenied, "access forbidden")
	}
	err = w.db.UpdateWorkoutExercise(rq.Exercise.WorkoutId, model.FromProto(rq.GetExercise()), rq.UpdateMask)
	if errors.Is(err, db.ErrorExerciseNotFound) {
		return nil, status.Error(codes.NotFound, "exercise not found")
	}
	if err != nil {
		log.Printf("error updating workout exercise: %v", err)
		return nil, status.Error(codes.Internal, "error updating workout exercise")
	}
	return &emptypb.Empty{}, nil
}

func (w *WorkoutAPI) ListWorkouts(context context.Context, _ *emptypb.Empty) (*workout.ListWorkoutsResponse, error) {
	userId, err := auth.GetUserId(context)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	workouts, err := w.db.GetWorkouts(userId)
	if err != nil {
		log.Printf("error listing workouts: %v", err)
		return nil, status.Error(codes.Internal, "error listing workouts")
	}
	var resp workout.ListWorkoutsResponse
	for _, wrk := range workouts {
		resp.Workouts = append(resp.Workouts, &workout.Workout{
			Id:   wrk.ID,
			Name: wrk.Name,
		})
	}
	return &resp, nil
}
