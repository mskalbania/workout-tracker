package api

import (
	"context"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"net"
	workout "proto/workout/v1/generated"
	"testing"
	"workout-tracker-server/mocks"
)

type ScheduleAPISuite struct {
	suite.Suite
	wsDbMock *mocks.WorkoutScheduleDb
	wDbMock  *mocks.WorkoutDb
	wsClient workout.WorkoutScheduleServiceClient
	cleanup  func()
}

func TestScheduleAPISuite(t *testing.T) {
	suite.Run(t, new(ScheduleAPISuite))
}

func (s *ScheduleAPISuite) SetupSuite() {
	wsDbMock := mocks.NewWorkoutScheduleDb(s.T())
	wDbMock := mocks.NewWorkoutDb(s.T())
	lis := bufconn.Listen(1024 * 1024)

	closeSrv := setupWorkoutScheduleTestServer(s.T(), lis, wDbMock, wsDbMock)
	client, closeCl := setupWorkoutScheduleTestClient(s.T(), lis)

	s.wsDbMock = wsDbMock
	s.wDbMock = wDbMock
	s.wsClient = client

	s.cleanup = func() {
		closeCl()
		closeSrv()
	}
}

func (s *ScheduleAPISuite) TearDownSuite() {
	s.cleanup()
}

func setupWorkoutScheduleTestServer(
	t *testing.T, listener *bufconn.Listener,
	wDbMock *mocks.WorkoutDb, wsDbMock *mocks.WorkoutScheduleDb,
) func() {
	server := grpc.NewServer()
	workout.RegisterWorkoutScheduleServiceServer(server, NewWorkoutScheduleAPI(wsDbMock, wDbMock))
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Fatalf("Server exited with error: %v", err)
		}
	}()
	return func() {
		server.Stop()
	}
}

func setupWorkoutScheduleTestClient(t *testing.T, listener *bufconn.Listener) (workout.WorkoutScheduleServiceClient, func()) {
	client, err := grpc.NewClient("passthrough://",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}
	return workout.NewWorkoutScheduleServiceClient(client), func() { client.Close() }
}
