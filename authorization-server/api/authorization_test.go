package api

import (
	"authorization-server/db"
	"authorization-server/mocks"
	"authorization-server/model"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"net"
	auth "proto/auth/v1/generated"
	"testing"
	"time"
)

var testUserName = "user1@gmail.com"
var testUserPassword = "password-password"
var testUserPasswordHash = "$2a$10$3.z9YPWhIeLbfXNzxLXSmePq4WtNidWrZl7pKLjoQP/WlNvp7yBa6"

type AuthorizationAPISuite struct {
	suite.Suite
	dbMock    *mocks.UserDb
	autClient auth.AuthorizationServiceClient
	cleanup   func()
}

func TestAuthorizationSuite(t *testing.T) {
	suite.Run(t, new(AuthorizationAPISuite))
}

func (s *AuthorizationAPISuite) SetupSuite() {
	dbMock := mocks.NewUserDb(s.T())
	lis := bufconn.Listen(1024 * 1024)

	closeSrv := setupServer(s.T(), lis, dbMock)
	client, closeCl := setupClient(s.T(), lis)

	s.dbMock = dbMock
	s.autClient = client

	s.cleanup = func() {
		closeCl()
		closeSrv()
	}
}

func setupServer(t *testing.T, listener *bufconn.Listener, dbMock *mocks.UserDb) func() {
	server := grpc.NewServer()
	auth.RegisterAuthorizationServiceServer(server, NewAuthorizationAPI(dbMock, JWTProperties{
		SigningKey:           []byte("secret"),
		AccessTokenDuration:  1,
		RefreshTokenDuration: 1,
	}, UTCTimeProvider{}))
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("error starting server: %v", err)
			return
		}
	}()
	return func() {
		server.Stop()
	}
}

func setupClient(t *testing.T, listener *bufconn.Listener) (auth.AuthorizationServiceClient, func()) {
	client, err := grpc.NewClient("passthrough://",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}
	return auth.NewAuthorizationServiceClient(client), func() { client.Close() }
}

func (s *AuthorizationAPISuite) TearDownSuite() {
	s.cleanup()
}

func (s *AuthorizationAPISuite) TestRegisterFailsUserAlreadyExists() {
	//given repository returns a user
	s.dbMock.EXPECT().Find(testUserName).Return(model.User{}, nil).Once()

	//when register is called
	rs, err := s.autClient.Register(context.Background(), &auth.RegisterRequest{
		Username: testUserName,
		Password: testUserPassword,
	})

	//then correct error is returned
	s.Require().Nil(rs)
	s.assertStatusError(codes.InvalidArgument, "user already exists", err)
}

func (s *AuthorizationAPISuite) TestRegisterFailsOnInternalError() {
	//given repository returns an error
	s.dbMock.EXPECT().Find(testUserName).Return(model.User{}, errors.New("some error")).Once()

	//when register is called
	rs, err := s.autClient.Register(context.Background(), &auth.RegisterRequest{
		Username: testUserName,
		Password: testUserPassword,
	})

	//then correct error is returned
	s.Require().Nil(rs)
	s.assertStatusError(codes.Internal, "error finding user", err)
}

func (s *AuthorizationAPISuite) TestRegisterFailsOnInvalidInput() {
	testCases := []struct {
		name    string
		request *auth.RegisterRequest
		errMsg  string
	}{
		{
			"EmptyUsername",
			&auth.RegisterRequest{Username: ""},
			"invalid RegisterRequest.Username: value must be a valid email address",
		},
		{
			"IncorrectEmailFormat",
			&auth.RegisterRequest{Username: "invalid"},
			"invalid RegisterRequest.Username: value must be a valid email address",
		},
		{
			"EmptyPassword",
			&auth.RegisterRequest{Username: testUserName, Password: ""},
			"invalid RegisterRequest.Password: value length must be between 10 and 25 runes, inclusive",
		},
		{
			"PasswordTooLong",
			&auth.RegisterRequest{Username: testUserName, Password: "qwertyuiopasdfghjklzxcvbnm"},
			"invalid RegisterRequest.Password: value length must be between 10 and 25 runes, inclusive",
		},
		{
			"PasswordTooHeavy",
			&auth.RegisterRequest{Username: testUserName, Password: "€€€€€€€€€€€€€€€€€€€€€€€€€"},
			"invalid RegisterRequest.Password: value length must be at most 70 bytes",
		},
	}
	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			//when register is called
			rs, err := s.autClient.Register(context.Background(), testCase.request)

			//then correct error is returned
			s.Require().Nil(rs)
			s.assertStatusError(codes.InvalidArgument, testCase.errMsg, err)
		})
	}
}

func (s *AuthorizationAPISuite) TestRegisterFailsOnSavingUserError() {
	//given no user found with given name
	s.dbMock.EXPECT().Find(testUserName).Return(model.User{}, db.ErrUserNotFound).Once()
	//and saving user fails
	s.dbMock.EXPECT().Save(mock.Anything).Return(model.User{}, errors.New("some error")).Once()

	//when register is called
	rs, err := s.autClient.Register(context.Background(), &auth.RegisterRequest{
		Username: testUserName,
		Password: testUserPassword,
	})

	//then correct error is returned
	s.Require().Nil(rs)
	s.assertStatusError(codes.Internal, "error saving user", err)
}

func (s *AuthorizationAPISuite) TestRegisterSuccess() {
	//given no user found with given name
	s.dbMock.EXPECT().Find(testUserName).Return(model.User{}, db.ErrUserNotFound).Once()
	//and saving user is successful
	s.dbMock.EXPECT().Save(mock.Anything).Return(model.User{ID: "id"}, nil).Once()

	//when register is called
	rs, err := s.autClient.Register(context.Background(), &auth.RegisterRequest{
		Username: testUserName,
		Password: testUserPassword,
	})

	//then correct response is returned
	s.Require().NoError(err)
	s.Require().NotNil(rs)
	s.EqualValues("id", rs.UserId)
}

func (s *AuthorizationAPISuite) TestLoginFailsOnUserNotFound() {
	//given repository returns an error
	s.dbMock.EXPECT().Find(testUserName).Return(model.User{}, db.ErrUserNotFound).Once()

	//when login is called
	rs, err := s.autClient.Login(context.Background(), &auth.LoginRequest{
		Username: testUserName,
	})

	//then correct error is returned
	s.Require().Nil(rs)
	s.assertStatusError(codes.Unauthenticated, "user not found", err)
}

func (s *AuthorizationAPISuite) TestLoginFailsOnInternalErrorFromDb() {
	//given repository returns an error
	s.dbMock.EXPECT().Find(testUserName).Return(model.User{}, errors.New("some error")).Once()

	//when login is called
	rs, err := s.autClient.Login(context.Background(), &auth.LoginRequest{
		Username: testUserName,
	})

	//then correct error is returned
	s.Require().Nil(rs)
	s.assertStatusError(codes.Internal, "error finding user", err)
}

func (s *AuthorizationAPISuite) TestLoginFailsOnInvalidPassword() {
	//given repository returns a user
	s.dbMock.EXPECT().Find(testUserName).Return(
		model.User{PasswordHash: testUserPasswordHash}, nil,
	).Once()

	//when login is called with invalid password
	rs, err := s.autClient.Login(context.Background(), &auth.LoginRequest{
		Username: testUserName,
		Password: "invalid",
	})

	//then correct error is returned
	s.Require().Nil(rs)
	s.assertStatusError(codes.Unauthenticated, "invalid credentials", err)
}

func (s *AuthorizationAPISuite) TestLoginSuccess() {
	//given repository returns a user
	s.dbMock.EXPECT().Find(testUserName).Return(
		model.User{PasswordHash: testUserPasswordHash}, nil,
	).Once()

	//when login is called with valid password
	rs, err := s.autClient.Login(context.Background(), &auth.LoginRequest{
		Username: testUserName,
		Password: testUserPassword,
	})

	//then correct response is returned
	s.Require().NoError(err)
	s.Require().NotNil(rs)
	s.NotEmpty(rs.Token)
}

func (s *AuthorizationAPISuite) assertStatusError(code codes.Code, message string, err error) {
	s.Require().NotNil(err, "Error is nil")
	st, ok := status.FromError(err)
	s.Require().True(ok, "Error is not a status error")
	s.Require().Equal(code, st.Code(), "Error code is not as expected - got: %v, expected: %v", st.Code(), code.String())
	s.Require().Equal(message, st.Message(), "Error message is incorrect")
}

func TestGenerateJWT(t *testing.T) {
	//given
	now := time.Now()
	fixedTimeProvider := mocks.NewTimeProvider(t)
	fixedTimeProvider.EXPECT().Now().Return(now).Twice()
	jwtProps := JWTProperties{
		SigningKey:          []byte("secret"),
		AccessTokenDuration: time.Minute,
	}

	//when
	token, err := generateJWT("user-id", jwtProps, fixedTimeProvider)

	//then
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtProps.SigningKey, nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)
	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	subject, err := claims.GetSubject()
	require.NoError(t, err)
	require.Equal(t, "user-id", subject)

	issuedAt, err := claims.GetIssuedAt()
	require.NoError(t, err)
	require.Equal(t, now.Unix(), issuedAt.Unix())

	expiresAt, err := claims.GetExpirationTime()
	require.NoError(t, err)
	require.Equal(t, now.Add(jwtProps.AccessTokenDuration).Unix(), expiresAt.Unix())
}
