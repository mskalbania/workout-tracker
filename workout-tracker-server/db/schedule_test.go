package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"workout-tracker-server/model"
	"workout-tracker-server/test"
)

type ScheduleSuite struct {
	suite.Suite
	wsDB              WorkoutScheduleDb
	existingWorkoutId string
	cleanup           func()
}

func TestScheduleSuite(t *testing.T) {
	suite.Run(t, new(ScheduleSuite))
}

func (s *ScheduleSuite) SetupSuite() {
	port, err, cleanup := test.SetupTestContainersDb()
	if err != nil {
		s.T().Fatal(err)
	}
	db := NewPostgresDb(fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres", port))
	workoutId, err := insertTestWorkout(db.db)
	if err != nil {
		s.T().Fatal(err)
	}
	s.existingWorkoutId = workoutId
	s.wsDB = db
	s.cleanup = cleanup
}

func insertTestWorkout(db *pgxpool.Pool) (string, error) {
	workoutId := uuid.New().String()
	_, err := db.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO workout (id, owner, name, comment) VALUES ('%s', '%s', 'test', 'test')",
			workoutId,
			uuid.New().String(),
		),
	)
	return workoutId, err
}

func (s *ScheduleSuite) TearDownSuite() {
	s.cleanup()
}

func (s *ScheduleSuite) TestSaveGetBetweenDates() {
	ws := model.WorkoutSchedule{
		OwnerID:     uuid.New().String(),
		WorkoutID:   s.existingWorkoutId,
		ScheduledAt: time.Now().UTC(),
	}

	id, err := s.wsDB.SaveWorkoutSchedule(ws)

	s.Require().NoError(err)
	s.Require().NotEmpty(id)

	wss, err := s.wsDB.GetWorkoutSchedulesBetweenDates(ws.OwnerID, time.Now().UTC().Add(-1*time.Hour), time.Now().UTC().Add(1*time.Hour))

	s.Require().NoError(err)
	s.Require().Len(wss, 1)
	s.Require().Equal(id, wss[0].ID)
	s.Require().Equal(ws.OwnerID, wss[0].OwnerID)
	s.Require().Equal(ws.WorkoutID, wss[0].WorkoutID)
	s.Require().Equal(ws.ScheduledAt, wss[0].ScheduledAt)
	s.Require().NotEmpty(wss[0].CreatedAt)
	s.Require().False(wss[0].Completed)
}

func (s *ScheduleSuite) TestGetWorkoutSchedulesBetweenDatesEmpty() {
	ws := model.WorkoutSchedule{
		OwnerID:     uuid.New().String(),
		WorkoutID:   s.existingWorkoutId,
		ScheduledAt: time.Now().UTC(),
	}

	_, err := s.wsDB.SaveWorkoutSchedule(ws)

	s.Require().NoError(err)

	//not within the time range
	wss, err := s.wsDB.GetWorkoutSchedulesBetweenDates(uuid.New().String(), time.Now().UTC().Add(1*time.Hour), time.Now().UTC().Add(2*time.Hour))

	s.Require().NoError(err)
	s.Require().Len(wss, 0)
}

func (s *ScheduleSuite) TestGetWorkoutSchedulesAtRangeBoundary() {
	scheduledAt := time.Now().UTC()
	testCases := []struct {
		name string
		from time.Time
		to   time.Time
	}{
		{
			"AtLeftBoundary",
			scheduledAt,
			scheduledAt.Add(1 * time.Hour),
		},
		{
			"AtRightBoundary",
			scheduledAt.Add(-1 * time.Hour),
			scheduledAt,
		},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			ws := model.WorkoutSchedule{
				OwnerID:     uuid.New().String(),
				WorkoutID:   s.existingWorkoutId,
				ScheduledAt: scheduledAt,
			}

			id, err := s.wsDB.SaveWorkoutSchedule(ws)

			require.NoError(t, err)
			require.NotEmpty(t, id)

			wss, err := s.wsDB.GetWorkoutSchedulesBetweenDates(ws.OwnerID, testCase.from, testCase.to)

			require.NoError(t, err)
			require.Len(t, wss, 1)
		})
	}
}

func (s *ScheduleSuite) TestUpdateCompleted() {
	ws := model.WorkoutSchedule{
		OwnerID:     uuid.New().String(),
		WorkoutID:   s.existingWorkoutId,
		ScheduledAt: time.Now().UTC(),
	}

	id, err := s.wsDB.SaveWorkoutSchedule(ws)

	s.Require().NoError(err)
	s.Require().NotEmpty(id)

	err = s.wsDB.UpdateWorkoutScheduleCompleted(id)

	s.Require().NoError(err)

	wss, err := s.wsDB.GetWorkoutSchedulesBetweenDates(ws.OwnerID, time.Now().UTC().Add(-1*time.Hour), time.Now().UTC().Add(1*time.Hour))

	s.Require().NoError(err)
	s.Require().Len(wss, 1)
	s.Require().True(wss[0].Completed)
}

func (s *ScheduleSuite) TestIsOwner() {
	owner := uuid.New().String()
	ws := model.WorkoutSchedule{
		OwnerID:     owner,
		WorkoutID:   s.existingWorkoutId,
		ScheduledAt: time.Now().UTC(),
	}

	id, err := s.wsDB.SaveWorkoutSchedule(ws)

	s.Require().NoError(err)

	isOwner, err := s.wsDB.IsWorkoutScheduleOwner(id, owner)

	s.Require().NoError(err)
	s.Require().True(isOwner)
}
