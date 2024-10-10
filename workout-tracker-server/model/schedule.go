package model

import (
	"time"
)

type WorkoutSchedule struct {
	ID          string
	OwnerID     string
	WorkoutID   string
	ScheduledAt time.Time
	CreatedAt   time.Time
	Completed   bool
}
