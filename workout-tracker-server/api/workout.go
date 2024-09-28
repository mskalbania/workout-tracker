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
	uuid, err := auth.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	isOwner, err := w.db.IsWorkoutOwner(rq.Workout.Id, uuid)
	if errors.Is(err, db.ErrWorkoutNotFound) {
		return nil, status.Error(codes.NotFound, "workout not found")
	}
	if err != nil {
		log.Printf("error getting workout data: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout data")
	}
	if !isOwner {
		return nil, status.Error(codes.PermissionDenied, "access forbidden")
	}
	err = w.db.UpdateWorkout(model.FromWorkoutProto(rq.GetWorkout()), rq.UpdateMask)
	if err != nil {
		log.Printf("error updating workout: %v", err)
		return nil, status.Error(codes.Internal, "error updating workout")
	}
	return &emptypb.Empty{}, nil
}

func (w *WorkoutAPI) UpdateWorkoutExercise(ctx context.Context, rq *workout.UpdateWorkoutExerciseRequest) (*emptypb.Empty, error) {
	uuid, err := auth.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	isOwner, err := w.db.IsWorkoutOwner(rq.Exercise.WorkoutId, uuid)
	if errors.Is(err, db.ErrWorkoutNotFound) {
		return nil, status.Error(codes.NotFound, "workout not found")
	}
	if err != nil {
		log.Printf("error getting workout data: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout data")
	}
	if !isOwner {
		return nil, status.Error(codes.PermissionDenied, "access forbidden")
	}
	err = w.db.UpdateWorkoutExercise(model.FromWorkoutExerciseProto(rq.GetExercise()), rq.UpdateMask)
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
		resp.Workouts = append(resp.Workouts, wrk.ToProto())
	}
	return &resp, nil
}

func (w *WorkoutAPI) GetWorkout(_ context.Context, rq *workout.GetWorkoutRequest) (*workout.GetWorkoutResponse, error) {
	wrk, err := w.db.GetWorkout(rq.Id)
	if errors.Is(err, db.ErrWorkoutNotFound) {
		return nil, status.Error(codes.NotFound, "workout not found")
	}
	if err != nil {
		log.Printf("error getting workout: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout")
	}
	return &workout.GetWorkoutResponse{
		Workout: wrk.ToProto(),
	}, nil
}
