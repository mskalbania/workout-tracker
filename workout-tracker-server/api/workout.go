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
	wrk := model.FromWorkoutProto(rq.Workout)
	wrk.OwnerID = userId
	id, err := w.db.SaveWorkout(wrk)
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

func (w *WorkoutAPI) UpdateWorkout(ctx context.Context, rq *workout.UpdateWorkoutRequest) (*emptypb.Empty, error) {
	if err := w.validateWorkoutOwner(ctx, rq.Workout.Id); err != nil {
		return nil, err
	}
	if err := w.db.UpdateWorkout(model.FromWorkoutProto(rq.GetWorkout()), rq.UpdateMask); err != nil {
		if errors.Is(err, db.ErrIncorrectExerciseReferenced) {
			return nil, status.Error(codes.InvalidArgument, "incorrect exercise referenced")
		}
		if errors.Is(err, db.ErrWorkoutExerciseNotFound) {
			return nil, status.Error(codes.NotFound, "workout exercise not found")
		}
		log.Printf("error updating workout: %v", err)
		return nil, status.Error(codes.Internal, "error updating workout")
	}
	return &emptypb.Empty{}, nil
}

func (w *WorkoutAPI) ListWorkouts(ctx context.Context, _ *emptypb.Empty) (*workout.ListWorkoutsResponse, error) {
	userId, err := auth.GetUserId(ctx)
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
		resp.Workouts = append(resp.Workouts, wrk.ToProto())
	}
	return &resp, nil
}

func (w *WorkoutAPI) GetWorkout(ctx context.Context, rq *workout.GetWorkoutRequest) (*workout.GetWorkoutResponse, error) {
	if err := w.validateWorkoutOwner(ctx, rq.Id); err != nil {
		return nil, err
	}
	if wrk, err := w.db.GetWorkout(rq.Id); err != nil {
		if errors.Is(err, db.ErrWorkoutNotFound) {
			return nil, status.Error(codes.NotFound, "workout not found")
		}
		log.Printf("error getting workout: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout")
	} else {
		return &workout.GetWorkoutResponse{
			Workout: wrk.ToProto(),
		}, nil
	}
}

func (w *WorkoutAPI) DeleteWorkout(ctx context.Context, rq *workout.DeleteWorkoutRequest) (*emptypb.Empty, error) {
	if err := w.validateWorkoutOwner(ctx, rq.Id); err != nil {
		return nil, err
	}
	if err := w.db.DeleteWorkout(rq.Id); err != nil {
		log.Printf("error deleting workout: %v", err)
		return nil, status.Error(codes.Internal, "error deleting workout")
	}
	return &emptypb.Empty{}, nil
}

func (w *WorkoutAPI) validateWorkoutOwner(ctx context.Context, workoutId string) error {
	userId, err := auth.GetUserId(ctx)
	if err != nil {
		return status.Error(codes.Internal, "user id not found in context")
	}
	isOwner, err := w.db.IsWorkoutOwner(workoutId, userId)
	if errors.Is(err, db.ErrWorkoutNotFound) {
		return status.Error(codes.NotFound, "workout not found")
	}
	if err != nil {
		log.Printf("error getting workout data: %v", err)
		return status.Error(codes.Internal, "error getting workout data")
	}
	if !isOwner {
		return status.Error(codes.PermissionDenied, "access forbidden")
	}
	return nil
}
