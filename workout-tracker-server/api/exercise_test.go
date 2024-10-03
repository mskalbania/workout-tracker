package api

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	workout "proto/workout/v1/generated"
	"testing"
	"workout-tracker-server/mocks"
	"workout-tracker-server/model"
)

type ExerciseAPISuite struct {
	suite.Suite
	dbMock         *mocks.ExerciseDb
	exerciseClient workout.ExerciseServiceClient
	cleanup        func()
}

func TestExerciseAPISuite(t *testing.T) {
	suite.Run(t, new(ExerciseAPISuite))
}

func (s *ExerciseAPISuite) SetupSuite() {
	dbMock := mocks.NewExerciseDb(s.T())
	lis := bufconn.Listen(1024 * 1024)

	closeSrv := setupServer(s.T(), lis, dbMock)
	client, closeCl := setupClient(s.T(), lis)

	s.dbMock = dbMock
	s.exerciseClient = client

	s.cleanup = func() {
		closeCl()
		closeSrv()
	}
}

func (s *ExerciseAPISuite) TearDownSuite() {
	s.cleanup()
}

func setupServer(t *testing.T, listener *bufconn.Listener, dbMock *mocks.ExerciseDb) func() {
	server := grpc.NewServer()
	workout.RegisterExerciseServiceServer(server, NewExerciseAPI(dbMock))
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	return func() {
		server.Stop()
	}
}

func setupClient(t *testing.T, listener *bufconn.Listener) (workout.ExerciseServiceClient, func()) {
	client, err := grpc.NewClient("passthrough://",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}
	return workout.NewExerciseServiceClient(client), func() { client.Close() }
}

func (s *ExerciseAPISuite) TestGetExercisesErrorFromDb() {
	s.dbMock.EXPECT().GetExercises("", "").Return(nil, fmt.Errorf("some err")).Once()

	resp, err := s.exerciseClient.GetExercises(context.Background(), &workout.GetExercisesRequest{})
	s.Require().Nil(resp)
	s.assertStatusError(codes.Internal, "error getting exercises", err)
}

func (s *ExerciseAPISuite) TestGetExercisesEmpty() {
	group := "group"
	s.dbMock.EXPECT().GetExercises(group, "").Return(nil, nil).Once()

	resp, err := s.exerciseClient.GetExercises(context.Background(), &workout.GetExercisesRequest{MuscleGroupFilter: &group})

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Empty(resp.Exercises)
}

func (s *ExerciseAPISuite) TestGetExercisesNonEmpty() {
	category := "category"
	s.dbMock.EXPECT().GetExercises("", category).Return([]model.Exercise{
		{
			ID:          "1",
			Name:        "exercise",
			MuscleGroup: "group",
			Category:    "category",
			Description: "description",
		},
	}, nil).Once()

	resp, err := s.exerciseClient.GetExercises(context.Background(), &workout.GetExercisesRequest{CategoryFilter: &category})

	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Exercises, 1)
	s.Require().Equal("1", resp.Exercises[0].Id)
	s.Require().Equal("exercise", resp.Exercises[0].Name)
	s.Require().Equal("group", resp.Exercises[0].MuscleGroup)
	s.Require().Equal("category", resp.Exercises[0].Category)
	s.Require().Equal("description", resp.Exercises[0].Description)
}

func (s *ExerciseAPISuite) assertStatusError(code codes.Code, message string, err error) {
	s.Require().NotNil(err, "Error is nil")
	st, ok := status.FromError(err)
	s.Require().True(ok, "Error is not a status error")
	s.Require().Equal(code, st.Code(), "Error code is not as expected - got: %v, expected: %v", st.Code(), code.String())
	s.Require().Equal(message, st.Message(), "Error message is incorrect")
}
