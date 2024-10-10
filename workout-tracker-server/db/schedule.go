package db

import (
	"context"
	"github.com/google/uuid"
	"time"
	"workout-tracker-server/model"
)

const (
	insertWorkoutSchedule          = "INSERT INTO workout_schedule(id, owner, workout, scheduled_at) VALUES ($1,$2,$3,$4)"
	updateWorkoutScheduleCompleted = "UPDATE workout_schedule SET completed = true WHERE id = $1"
	selectWorkoutScheduleOwner     = "SELECT owner FROM workout_schedule WHERE id = $1"
	selectWorkoutSchedulesBetween  = "SELECT id, owner, workout, scheduled_at, crated_at, completed FROM workout_schedule WHERE owner = $1 AND crated_at >= $2 AND crated_at <= $3"
)

type WorkoutScheduleDb interface {
	SaveWorkoutSchedule(ws model.WorkoutSchedule) (string, error)
	UpdateWorkoutScheduleCompleted(scheduleId string) error
	IsWorkoutScheduleOwner(scheduleId, userId string) (bool, error)
	GetWorkoutSchedulesBetweenDates(userId string, from, to time.Time) ([]model.WorkoutSchedule, error)
}

func (p *PostgresDb) SaveWorkoutSchedule(ws model.WorkoutSchedule) (string, error) {
	id := uuid.New().String()
	_, err := p.db.Exec(context.Background(), insertWorkoutSchedule, id, ws.OwnerID, ws.WorkoutID, ws.ScheduledAt)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *PostgresDb) UpdateWorkoutScheduleCompleted(scheduleId string) error {
	_, err := p.db.Exec(context.Background(), updateWorkoutScheduleCompleted, scheduleId)
	return err
}

func (p *PostgresDb) IsWorkoutScheduleOwner(scheduleId, userId string) (bool, error) {
	var owner string
	err := p.db.QueryRow(context.Background(), selectWorkoutScheduleOwner, scheduleId).Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner == userId, nil
}

func (p *PostgresDb) GetWorkoutSchedulesBetweenDates(userId string, from, to time.Time) ([]model.WorkoutSchedule, error) {
	rows, err := p.db.Query(context.Background(), selectWorkoutSchedulesBetween, userId, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []model.WorkoutSchedule
	for rows.Next() {
		var ws model.WorkoutSchedule
		err := rows.Scan(&ws.ID, &ws.OwnerID, &ws.WorkoutID, &ws.ScheduledAt, &ws.CreatedAt, &ws.Completed)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, ws)
	}
	return schedules, nil
}
