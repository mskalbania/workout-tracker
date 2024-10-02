package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
	"workout-tracker-server/model"
)

type ExerciseSuite struct {
	suite.Suite
	exerciseDb ExerciseDb
	cleanup    func()
}

func TestExerciseSuite(t *testing.T) {
	suite.Run(t, new(ExerciseSuite))
}

func (s *ExerciseSuite) SetupSuite() {
	pgCt, err := postgres.Run(context.Background(),
		"postgres:16-alpine",
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithDatabase("postgres"),
		postgres.WithInitScripts("../../init.sql"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		s.T().Fatal(err)
	}
	port, err := pgCt.MappedPort(context.Background(), "5432")
	if err != nil {
		s.T().Fatal(err)
	}
	s.T().Log("Started postgres container on port: ", port.Int())
	s.exerciseDb = NewPostgresDb(fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres", port.Int()))
	s.cleanup = func() {
		pgCt.Terminate(context.Background())
	}
}

func (s *ExerciseSuite) TearDownSuite() {
	s.cleanup()
}

func (s *ExerciseSuite) TestGetExercises() {
	testData := []struct {
		name              string
		muscleGroupQuery  string
		categoryQuery     string
		expectedExercises []model.Exercise
	}{
		{
			name:              "NoExerciseMatch",
			muscleGroupQuery:  "non-existent",
			expectedExercises: nil,
		},
		{
			name:             "MuscleGroupMatch",
			muscleGroupQuery: "legs",
			expectedExercises: []model.Exercise{
				{
					ID:          "c3339fa8-f9d6-481d-b983-f9cdc24ca4d0",
					Name:        "Squat",
					Description: "The squat is a lower body exercise.",
					MuscleGroup: "LEGS",
					Category:    "STRENGTH",
				},
			},
		},
		{
			name:          "CategoryMatch",
			categoryQuery: "Cardio",
			expectedExercises: []model.Exercise{
				{
					ID:          "94b4109b-25ba-4519-8aa7-6adef75c0d37",
					Name:        "Pull-up",
					Description: "A pull-up is an upper-body strength exercise.",
					MuscleGroup: "BACK",
					Category:    "CARDIO",
				},
			},
		},
		{
			name:             "MuscleGroupAndCategoryMatch",
			muscleGroupQuery: "BACK",
			categoryQuery:    "Cardio",
			expectedExercises: []model.Exercise{
				{
					ID:          "94b4109b-25ba-4519-8aa7-6adef75c0d37",
					Name:        "Pull-up",
					Description: "A pull-up is an upper-body strength exercise.",
					MuscleGroup: "BACK",
					Category:    "CARDIO",
				},
			},
		},
	}
	for _, tt := range testData {
		s.Run(tt.name, func() {
			exercises, err := s.exerciseDb.GetExercises(tt.muscleGroupQuery, tt.categoryQuery)
			s.Require().NoError(err)
			s.Require().Equal(tt.expectedExercises, exercises)
		})
	}
}

func (s *ExerciseSuite) TestGetExercisesNoQuery() {
	exercises, err := s.exerciseDb.GetExercises("", "")
	s.Require().NoError(err)
	s.Require().Len(exercises, 4)
}
