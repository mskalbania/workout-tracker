package api

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	workout "proto/workout/v1/generated"
	"workout-tracker-server/auth"
	"workout-tracker-server/db"
	"workout-tracker-server/model"
)

type WorkoutScheduleAPI struct {
	workout.UnimplementedWorkoutScheduleServiceServer
	wsDb db.WorkoutScheduleDb
	wDb  db.WorkoutDb
}

func NewWorkoutScheduleAPI(wsDb db.WorkoutScheduleDb, wDb db.WorkoutDb) *WorkoutScheduleAPI {
	return &WorkoutScheduleAPI{wsDb: wsDb, wDb: wDb}
}

func (s *WorkoutScheduleAPI) ScheduleWorkout(ctx context.Context, rq *workout.ScheduleWorkoutRequest) (*workout.ScheduleWorkoutResponse, error) {
	if err := rq.Validate(); err != nil {
		return nil, validationError(func() error { return err.(workout.ScheduleWorkoutRequestValidationError).Cause() })
	}
	userId, err := s.getValidatedWorkoutOwnerId(ctx, rq.WorkoutSchedule.WorkoutId)
	if err != nil {
		return nil, err
	}
	ws := model.WorkoutSchedule{
		OwnerID:     userId,
		WorkoutID:   rq.WorkoutSchedule.WorkoutId,
		ScheduledAt: rq.WorkoutSchedule.ScheduleAt.AsTime(),
	}
	id, err := s.wsDb.SaveWorkoutSchedule(ws)
	if err != nil {
		log.Printf("error saving workout schedule: %v", err)
		return nil, status.Error(codes.Internal, "error saving workout schedule")
	}
	return &workout.ScheduleWorkoutResponse{
		Id: id,
	}, nil
}

func (s *WorkoutScheduleAPI) MarkWorkoutComplete(ctx context.Context, rq *workout.MarkWorkoutCompleteRequest) (*emptypb.Empty, error) {
	if err := rq.Validate(); err != nil {
		return nil, validationError(func() error { return err.(workout.MarkWorkoutCompleteRequestValidationError) })
	}
	userId, err := auth.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	isOwner, err := s.wsDb.IsWorkoutScheduleOwner(rq.Id, userId)
	if err != nil {
		log.Printf("error getting workout schedule owner: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout schedule owner")
	}
	if !isOwner {
		return nil, status.Error(codes.PermissionDenied, "access forbidden")
	}
	err = s.wsDb.UpdateWorkoutScheduleCompleted(rq.Id)
	if err != nil {
		log.Printf("error updating workout schedule: %v", err)
		return nil, status.Error(codes.Internal, "error updating workout schedule")
	}
	return &emptypb.Empty{}, nil
}

func (s *WorkoutScheduleAPI) GetWorkoutScheduleReport(ctx context.Context, rq *workout.GetWorkoutScheduleReportRequest) (*workout.GetWorkoutScheduleReportResponse, error) {
	if err := rq.Validate(); err != nil {
		return nil, validationError(func() error { return err.(workout.GetWorkoutScheduleReportRequestValidationError) })
	}
	userId, err := auth.GetUserId(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}
	schedules, err := s.wsDb.GetWorkoutSchedulesBetweenDates(userId, rq.StartDate.AsTime(), rq.EndDate.AsTime())
	if err != nil {
		log.Printf("error getting workout schedules: %v", err)
		return nil, status.Error(codes.Internal, "error getting workout schedules")
	}
	var respSchedules []*workout.WorkoutSchedule
	for _, ws := range schedules {
		respSchedules = append(respSchedules, &workout.WorkoutSchedule{
			Id:         ws.ID,
			WorkoutId:  ws.WorkoutID,
			ScheduleAt: timestamppb.New(ws.ScheduledAt),
			CreatedAt:  timestamppb.New(ws.CreatedAt),
			Completed:  ws.Completed,
		})
	}
	return &workout.GetWorkoutScheduleReportResponse{
		WorkoutSchedules: respSchedules,
	}, nil
}

func (s *WorkoutScheduleAPI) getValidatedWorkoutOwnerId(ctx context.Context, workoutId string) (string, error) {
	userId, err := auth.GetUserId(ctx)
	if err != nil {
		return "", status.Error(codes.Internal, "user id not found in context")
	}
	isOwner, err := s.wDb.IsWorkoutOwner(workoutId, userId)
	if errors.Is(err, db.ErrWorkoutNotFound) {
		return "", status.Error(codes.NotFound, "workout not found")
	}
	if err != nil {
		log.Printf("error getting workout data: %v", err)
		return "", status.Error(codes.Internal, "error getting workout data")
	}
	if !isOwner {
		return "", status.Error(codes.PermissionDenied, "access forbidden")
	}
	return userId, nil
}
