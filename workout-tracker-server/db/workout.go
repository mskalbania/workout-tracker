package db

import "workout-tracker-server/model"

type WorkoutDb interface {
	CreateWorkout(workout model.Workout)
}

func (p *PostgresDb) CreateWorkout(workout model.Workout) {

}
