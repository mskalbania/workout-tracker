package model

type Workout struct {
	ID        string
	Owner     string
	Name      string
	Exercises []Exercise
}
