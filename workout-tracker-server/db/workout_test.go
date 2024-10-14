package db

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"testing"
	"workout-tracker-server/model"
	"workout-tracker-server/test"
)

const (
	existingExerciseId    = "87df312d-36e0-40e8-915e-093ac3342ac8"
	existingExerciseId2   = "c3339fa8-f9d6-481d-b983-f9cdc24ca4d0"
	nonExistingExerciseId = "a51e4f9b-ee5a-4d11-ba8b-100941daf00f"
)

type WorkoutSuite struct {
	suite.Suite
	workoutDb WorkoutDb
	cleanup   func()
}

func TestWorkoutSuite(t *testing.T) {
	suite.Run(t, new(WorkoutSuite))
}

func (s *WorkoutSuite) SetupSuite() {
	port, err, cleanup := test.SetupTestContainersDb()
	if err != nil {
		s.T().Fatal(err)
	}
	s.workoutDb = NewPostgresDb(fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres", port))
	s.cleanup = cleanup
}

func (s *WorkoutSuite) TearDownSuite() {
	s.cleanup()
}

func (s *WorkoutSuite) TestSaveGetWorkoutSuccessful() {
	userId := uuid.New().String()
	comment := "Some comment"
	weight := int32(20)
	testData := []struct {
		name    string
		workout model.Workout
	}{
		{
			name: "WorkoutWithNoExercises",
			workout: model.Workout{
				OwnerID: userId,
				Name:    "Some workout",
				Comment: &comment,
			},
		},
		{
			name: "WorkoutWithExercises",
			workout: model.Workout{
				OwnerID: userId,
				Name:    "Some workout",
				Exercises: []model.WorkoutExercise{
					{ //one exercise with all fields
						ExerciseID:  existingExerciseId,
						Order:       1,
						Repetitions: 10,
						Sets:        3,
						Weight:      &weight,
						Comment:     &comment,
					},
					{ //second exercise with only required fields
						ExerciseID:  existingExerciseId,
						Order:       2,
						Repetitions: 10,
						Sets:        3,
					},
				},
			},
		},
	}
	for _, tCase := range testData {
		s.T().Run(tCase.name, func(t *testing.T) {
			workoutId, err := s.workoutDb.SaveWorkout(tCase.workout)

			s.Require().NoError(err)
			s.Require().NotEmpty(workoutId)

			workout, err := s.workoutDb.GetWorkout(workoutId)

			s.Require().NoError(err)
			s.Require().Equal(tCase.workout.OwnerID, workout.OwnerID)
			s.Require().Equal(tCase.workout.Name, workout.Name)
			s.Require().Equal(tCase.workout.Comment, workout.Comment)
			s.Require().Equal(len(tCase.workout.Exercises), len(workout.Exercises))
			for i, expectedEx := range tCase.workout.Exercises {
				s.Require().Equal(expectedEx.ExerciseID, workout.Exercises[i].ExerciseID)
				s.Require().Equal(expectedEx.Order, workout.Exercises[i].Order)
				s.Require().Equal(expectedEx.Repetitions, workout.Exercises[i].Repetitions)
				s.Require().Equal(expectedEx.Sets, workout.Exercises[i].Sets)
				s.Require().Equal(expectedEx.Weight, workout.Exercises[i].Weight)
				s.Require().Equal(expectedEx.Comment, workout.Exercises[i].Comment)
			}
		})
	}
}

func (s *WorkoutSuite) TestSaveIsTransactional() {
	//given workout that references non-existing exercise
	userId := uuid.New().String()
	workout := model.Workout{
		OwnerID: userId,
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID: nonExistingExerciseId,
			},
		},
	}

	//when saving workout
	_, err := s.workoutDb.SaveWorkout(workout)

	//then
	s.Require().Error(err)

	//and when getting workout
	workouts, err := s.workoutDb.GetWorkouts(userId)

	//then partial workout data is not saved
	s.Require().NoError(err)
	s.Require().Len(workouts, 0)
}

func (s *WorkoutSuite) TestGetWorkoutsEmpty() {
	//given user that has no registered workouts
	userId := uuid.New().String()

	//when
	wrks, err := s.workoutDb.GetWorkouts(userId)

	//then
	s.Require().NoError(err)
	s.Require().Len(wrks, 0)
}

func (s *WorkoutSuite) TestGetWorkoutsNonEmpty() {
	//given user have some workouts registered
	comment := "Comment"
	userid := uuid.New().String()
	wkrId1, err := s.workoutDb.SaveWorkout(model.Workout{
		OwnerID: userid,
		Name:    "WRK1",
		Comment: &comment,
	})
	s.Require().NoError(err)
	wrkId2, err := s.workoutDb.SaveWorkout(model.Workout{
		OwnerID: userid,
		Name:    "WRK2",
	})
	s.Require().NoError(err)

	//when
	wrks, err := s.workoutDb.GetWorkouts(userid)

	//then
	s.Require().NoError(err)
	s.Require().Len(wrks, 2)

	s.Require().Equal(wkrId1, wrks[0].ID)
	s.Require().Equal("WRK1", wrks[0].Name)
	s.Require().Equal(&comment, wrks[0].Comment)

	s.Require().Equal(wrkId2, wrks[1].ID)
	s.Require().Equal("WRK2", wrks[1].Name)
	s.Require().Nil(wrks[1].Comment)
}

func (s *WorkoutSuite) TestDeleteWorkoutNonExisting() {
	//when
	err := s.workoutDb.DeleteWorkout(uuid.New().String())

	//then
	s.Require().NoError(err)
}

func (s *WorkoutSuite) TestDeleteWorkoutExisting() {
	//given
	userId := uuid.New().String()
	wrkId, err := s.workoutDb.SaveWorkout(model.Workout{
		OwnerID: userId,
		Name:    "WRK",
	})
	s.Require().NoError(err)

	//when
	err = s.workoutDb.DeleteWorkout(wrkId)

	//then
	s.Require().NoError(err)

	//and when
	wrks, err := s.workoutDb.GetWorkouts(userId)

	//then
	s.Require().NoError(err)
	s.Require().Len(wrks, 0)
}

func (s *WorkoutSuite) TestIsWorkoutOwnerWorkoutExists() {
	//given
	userId := uuid.New().String()
	testData := []struct {
		name            string
		workout         model.Workout
		userId          string
		expectedOutcome bool
	}{
		{
			name: "IsOwner",
			workout: model.Workout{
				OwnerID: userId,
			},
			userId:          userId,
			expectedOutcome: true,
		},
		{
			name: "IsNotOwner",
			workout: model.Workout{
				OwnerID: uuid.New().String(),
			},
			userId:          uuid.New().String(),
			expectedOutcome: false,
		},
	}

	for _, tCase := range testData {
		s.T().Run(tCase.name, func(t *testing.T) {
			workoutId, err := s.workoutDb.SaveWorkout(tCase.workout)
			s.Require().NoError(err)

			//when
			isOwner, err := s.workoutDb.IsWorkoutOwner(workoutId, tCase.userId)

			//then
			s.Require().NoError(err)
			s.Require().Equal(tCase.expectedOutcome, isOwner)
		})
	}
}

func (s *WorkoutSuite) TestIsWorkoutOwnerWorkoutNotExists() {
	//when
	isOwner, err := s.workoutDb.IsWorkoutOwner(uuid.New().String(), uuid.New().String())

	//then
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrWorkoutNotFound)
	s.Require().False(isOwner)
}

func (s *WorkoutSuite) TestGetWorkoutNotExists() {
	//when
	_, err := s.workoutDb.GetWorkout(uuid.New().String())

	//then
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrWorkoutNotFound)
}

func (s *WorkoutSuite) TestUpdateWorkoutNotExists() {
	//when
	err := s.workoutDb.UpdateWorkout(model.Workout{
		ID: uuid.New().String(),
	}, &fieldmaskpb.FieldMask{Paths: []string{"name"}})

	//then
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrWorkoutNotFound)
}

func (s *WorkoutSuite) TestUpdateWorkoutUsesMasks() {
	comment := "Comment"
	comment2 := "Comment2"
	weight1 := int32(20)
	weight2 := int32(30)
	testData := []struct {
		name           string
		initialWorkout model.Workout
		mask           *fieldmaskpb.FieldMask
		updatedWorkout model.Workout
	}{
		{
			name: "NoFieldsToUpdate",
			initialWorkout: model.Workout{
				Name:      "WRK",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb28",
				Comment:   &comment,
				Exercises: []model.WorkoutExercise{},
			},
			mask: &fieldmaskpb.FieldMask{},
			updatedWorkout: model.Workout{
				Name:      "WRK",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb28",
				Comment:   &comment,
				Exercises: []model.WorkoutExercise{},
			},
		},
		{
			name: "UpdateWorkoutName",
			initialWorkout: model.Workout{
				Name:      "WRK",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb27",
				Comment:   &comment,
				Exercises: []model.WorkoutExercise{},
			},
			mask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			updatedWorkout: model.Workout{
				Name:      "WRK2",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb27",
				Comment:   &comment,
				Exercises: []model.WorkoutExercise{},
			},
		},
		{
			name: "UpdateWorkoutComment",
			initialWorkout: model.Workout{
				Name:      "WRK",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb26",
				Comment:   &comment,
				Exercises: []model.WorkoutExercise{},
			},
			mask: &fieldmaskpb.FieldMask{Paths: []string{"comment"}},
			updatedWorkout: model.Workout{
				Name:      "WRK",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb26",
				Comment:   &comment2,
				Exercises: []model.WorkoutExercise{},
			},
		},
		{
			name: "UpdateWorkoutNameComment",
			initialWorkout: model.Workout{
				Name:      "WRK",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb29",
				Comment:   &comment,
				Exercises: []model.WorkoutExercise{},
			},
			mask: &fieldmaskpb.FieldMask{Paths: []string{"name", "comment"}},
			updatedWorkout: model.Workout{
				Name:      "WRK2",
				OwnerID:   "c6278149-3c02-4770-8e8f-fabb1c31eb29",
				Comment:   &comment2,
				Exercises: []model.WorkoutExercise{},
			},
		},
		{
			name: "UpdateWorkoutExercises",
			initialWorkout: model.Workout{
				Name:    "WRK",
				OwnerID: "c6278149-3c02-4770-8e8f-fabb1c31eb33",
				Comment: &comment,
				Exercises: []model.WorkoutExercise{
					{
						ExerciseID:  existingExerciseId,
						Order:       1,
						Repetitions: 10,
						Sets:        3,
						Weight:      &weight1,
						Comment:     &comment,
					},
				},
			},
			mask: &fieldmaskpb.FieldMask{Paths: []string{"exercises"}},
			updatedWorkout: model.Workout{
				Name:    "WRK",
				OwnerID: "c6278149-3c02-4770-8e8f-fabb1c31eb33",
				Comment: &comment,
				Exercises: []model.WorkoutExercise{
					{
						ExerciseID:  existingExerciseId2,
						Order:       2,
						Repetitions: 11,
						Sets:        4,
						Weight:      &weight2,
						Comment:     &comment2,
					},
				},
			},
		},
	}
	for _, tCase := range testData {
		s.T().Run(tCase.name, func(t *testing.T) {
			//given
			wrkId, err := s.workoutDb.SaveWorkout(tCase.initialWorkout)
			s.Require().NoError(err)
			tCase.updatedWorkout.ID = wrkId

			//when
			err = s.workoutDb.UpdateWorkout(tCase.updatedWorkout, tCase.mask)

			//then
			s.Require().NoError(err)

			//and when
			wrk, err := s.workoutDb.GetWorkout(wrkId)

			//then
			s.Require().NoError(err)
			s.Require().NoError(err)
			s.Require().Equal(tCase.updatedWorkout.OwnerID, wrk.OwnerID)
			s.Require().Equal(tCase.updatedWorkout.Name, wrk.Name)
			s.Require().Equal(tCase.updatedWorkout.Comment, wrk.Comment)
			s.Require().Equal(len(tCase.updatedWorkout.Exercises), len(wrk.Exercises))
			for i, expectedEx := range tCase.updatedWorkout.Exercises {
				s.Require().Equal(expectedEx.ExerciseID, wrk.Exercises[i].ExerciseID)
				s.Require().Equal(expectedEx.Order, wrk.Exercises[i].Order)
				s.Require().Equal(expectedEx.Repetitions, wrk.Exercises[i].Repetitions)
				s.Require().Equal(expectedEx.Sets, wrk.Exercises[i].Sets)
				s.Require().Equal(expectedEx.Weight, wrk.Exercises[i].Weight)
				s.Require().Equal(expectedEx.Comment, wrk.Exercises[i].Comment)
			}
		})
	}
}

func (s *WorkoutSuite) TestUpdateIsTransactional() {
	//given workout that references non-existing exercise
	userId := uuid.New().String()
	workout := model.Workout{
		Name:    "WRK",
		OwnerID: userId,
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID: existingExerciseId,
			},
		},
	}
	workoutId, err := s.workoutDb.SaveWorkout(workout)
	s.Require().NoError(err)

	//when updating workout
	err = s.workoutDb.UpdateWorkout(model.Workout{
		ID:   workoutId,
		Name: "WRK2",
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID: nonExistingExerciseId,
			},
		},
	}, &fieldmaskpb.FieldMask{Paths: []string{"exercises"}})

	//then
	s.Require().Error(err)

	//and when getting workout
	wrk, err := s.workoutDb.GetWorkout(workoutId)

	//then name not updated - no partial update
	s.Require().NoError(err)
	s.Require().Equal(workout.Name, wrk.Name)
}

func (s *WorkoutSuite) TestUpdateWorkoutReferencesNonExistingWorkoutExercise() {
	//given
	userId := uuid.New().String()
	workoutId, err := s.workoutDb.SaveWorkout(model.Workout{
		Name:    "WRK",
		OwnerID: userId,
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID: existingExerciseId,
			},
		},
	})
	s.Require().NoError(err)

	//when
	err = s.workoutDb.UpdateWorkout(model.Workout{
		ID:   workoutId,
		Name: "WRK2",
		Exercises: []model.WorkoutExercise{
			{
				WorkoutExerciseID: uuid.New().String(),
			},
		},
	}, &fieldmaskpb.FieldMask{Paths: []string{"exercises"}})

	//then
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrWorkoutExerciseNotFound)
}

func (s *WorkoutSuite) TestUpdateWorkoutSaveNewWorkoutExercise() {
	//given
	userId := uuid.New().String()
	workoutId, err := s.workoutDb.SaveWorkout(model.Workout{
		Name:      "WRK",
		OwnerID:   userId,
		Exercises: []model.WorkoutExercise{},
	})
	s.Require().NoError(err)

	//when
	err = s.workoutDb.UpdateWorkout(model.Workout{
		ID: workoutId,
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID: existingExerciseId,
			},
		},
	}, &fieldmaskpb.FieldMask{Paths: []string{"exercises"}})

	//then
	s.Require().NoError(err)

	wrk, err := s.workoutDb.GetWorkout(workoutId)
	s.Require().NoError(err)
	s.Require().Len(wrk.Exercises, 1)
	s.Require().NotEmpty(wrk.Exercises[0].WorkoutExerciseID)
}

func (s *WorkoutSuite) TestUpdateWorkoutUpdateExistingWorkoutExercise() {
	//given
	userId := uuid.New().String()
	comment1 := "Comment"
	comment2 := "Comment2"
	weight1 := int32(20)
	weight2 := int32(30)
	workoutId, err := s.workoutDb.SaveWorkout(model.Workout{
		Name:    "WRK",
		OwnerID: userId,
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID:  existingExerciseId,
				Repetitions: 10,
				Order:       1,
				Sets:        3,
				Weight:      &weight1,
				Comment:     &comment1,
			},
		},
	})
	s.Require().NoError(err)

	wrk, err := s.workoutDb.GetWorkout(workoutId)
	s.Require().NoError(err)
	s.Require().Len(wrk.Exercises, 1)
	workoutExerciseId := wrk.Exercises[0].WorkoutExerciseID

	//when
	err = s.workoutDb.UpdateWorkout(model.Workout{
		ID: workoutId,
		Exercises: []model.WorkoutExercise{
			{
				WorkoutExerciseID: workoutExerciseId,
				ExerciseID:        existingExerciseId2,
				Repetitions:       11,
				Order:             2,
				Sets:              4,
				Weight:            &weight2,
				Comment:           &comment2,
			},
		},
	}, &fieldmaskpb.FieldMask{Paths: []string{"exercises"}})

	//then
	s.Require().NoError(err)

	wrk2, err := s.workoutDb.GetWorkout(workoutId)
	s.Require().NoError(err)

	s.Require().Len(wrk.Exercises, 1)
	s.Require().Equal(workoutExerciseId, wrk2.Exercises[0].WorkoutExerciseID)
	s.Require().Equal(existingExerciseId2, wrk2.Exercises[0].ExerciseID)
	s.Require().Equal(int32(11), wrk2.Exercises[0].Repetitions)
	s.Require().Equal(int32(2), wrk2.Exercises[0].Order)
	s.Require().Equal(int32(4), wrk2.Exercises[0].Sets)
	s.Require().Equal(&weight2, wrk2.Exercises[0].Weight)
	s.Require().Equal(&comment2, wrk2.Exercises[0].Comment)
}

func (s *WorkoutSuite) TestUpdateWorkoutDeleteExistingWorkoutExercise() {
	//given
	userId := uuid.New().String()
	workoutId, err := s.workoutDb.SaveWorkout(model.Workout{
		Name:    "WRK",
		OwnerID: userId,
		Exercises: []model.WorkoutExercise{
			{
				ExerciseID: existingExerciseId,
			},
		},
	})
	s.Require().NoError(err)

	//when
	err = s.workoutDb.UpdateWorkout(model.Workout{
		ID:        workoutId,
		Exercises: []model.WorkoutExercise{},
	}, &fieldmaskpb.FieldMask{Paths: []string{"exercises"}})

	//then
	s.Require().NoError(err)

	wrk2, err := s.workoutDb.GetWorkout(workoutId)
	s.Require().NoError(err)

	s.Require().Len(wrk2.Exercises, 0)
}
