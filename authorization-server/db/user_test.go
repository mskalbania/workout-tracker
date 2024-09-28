package db

import (
	"authorization-server/model"
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"testing"
)

type UserDbSuite struct {
	suite.Suite
	userDb  UserDb
	cleanup func()
}

func TestUserDbSuite(t *testing.T) {
	suite.Run(t, new(UserDbSuite))
}

func (s *UserDbSuite) SetupSuite() {
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
	s.userDb = NewPostgresDb(fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres", port.Int()))
	s.cleanup = func() {
		pgCt.Terminate(context.Background())
	}
}

func (s *UserDbSuite) TearDownSuite() {
	s.cleanup()
}

func (s *UserDbSuite) TestSaveFindUser() {
	//given
	user := model.User{
		Username:     "user@gmail.com",
		PasswordHash: "hash",
	}

	//when
	savedUser, err := s.userDb.Save(user)

	//then
	s.Require().NoError(err)
	s.Require().NotNil(savedUser)
	s.Require().NotEmpty(savedUser.ID)

	//and when
	foundUser, err := s.userDb.Find(user.Username)

	//then
	s.Require().NoError(err)
	s.Require().NotNil(foundUser)
	s.Require().Equal(user.Username, foundUser.Username)
	s.Require().Equal(user.PasswordHash, foundUser.PasswordHash)
}

func (s *UserDbSuite) TestFindUserNotFound() {
	//when
	user, err := s.userDb.Find("not-found")

	//then
	s.Require().Error(err)
	s.Require().Equal(ErrUserNotFound, err)
	s.Require().Empty(user)
}
