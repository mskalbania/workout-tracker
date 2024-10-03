package model

import workout "proto/workout/v1/generated"

type Exercise struct {
	ID          string
	Name        string
	Description string
	MuscleGroup string
	Category    string
}

func (e Exercise) ToProto() *workout.Exercise {
	return &workout.Exercise{
		Id:          e.ID,
		Name:        e.Name,
		MuscleGroup: e.MuscleGroup,
		Category:    e.Category,
		Description: e.Description,
	}
}
