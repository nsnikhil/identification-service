package session_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/liberr"
	"identification-service/pkg/session"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
	"testing"
	"time"
)

//TODO: FIX LINE BREAK ON THIS FILE
func TestLoginUserSuccess(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, userID, map[string]string{"session_id": "f113fe5c-de2f-4876-b734-b51fbdc96e4b"}).Return(accessToken, nil)
	mockGenerator.On("GenerateRefreshToken").Return(refreshToken, nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

	mockClientService := &client.MockService{}
	mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 14440, nil)

	service := session.NewService(mockStore, mockUserService, mockClientService, mockGenerator)

	_, _, err := service.LoginUser(context.Background(), clientName, clientSecret, email, userPassword)
	require.NoError(t, err)
}

func TestLoginUserFailure(t *testing.T) {
	testCases := map[string]struct {
		store         func() session.Store
		userService   func() user.Service
		clientService func() client.Service
		generator     func() token.Generator
	}{
		"test failure get user id fails": {
			store: func() session.Store { return &session.MockStore{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return("", liberr.WithArgs(errors.New("failed to get user id")))

				return mockUserService
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator:     func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure get refresh token generation fails": {
			store: func() session.Store { return &session.MockStore{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

				return mockUserService
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return("", liberr.WithArgs(errors.New("failed to generate refresh token")))

				return mockGenerator
			},
		},
		"test failure when session creation fails due to invalid user id": {
			store: func() session.Store { return &session.MockStore{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return("invalidUserID", nil)

				return mockUserService
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(refreshToken, nil)

				return mockGenerator
			},
		},
		"test failure when session creation fails due to invalid refresh token": {
			store: func() session.Store { return &session.MockStore{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

				return mockUserService
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return("invalidRefreshToken", nil)

				return mockGenerator
			},
		},
		"test failure when store call fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Session")).Return("", liberr.WithArgs(errors.New("failed to create new session")))

				return mockStore
			},
			clientService: func() client.Service { return &client.MockService{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(refreshToken, nil)

				return mockGenerator
			},
		},
		"test failure when get client ttl fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

				return mockUserService
			},
			clientService: func() client.Service {
				mockClientService := &client.MockService{}
				mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(-1, -1, liberr.WithArgs(errors.New("failed to get client ttl")))

				return mockClientService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(refreshToken, nil)

				return mockGenerator
			},
		},
		"test failure when access token generation fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

				return mockUserService
			},
			clientService: func() client.Service {
				mockClientService := &client.MockService{}
				mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 14440, nil)

				return mockClientService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(refreshToken, nil)
				mockGenerator.On("GenerateAccessToken", 10, userID, map[string]string{"session_id": "f113fe5c-de2f-4876-b734-b51fbdc96e4b"}).Return("", liberr.WithArgs(errors.New("failed to generate access token")))

				return mockGenerator
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := session.NewService(testCase.store(), testCase.userService(), testCase.clientService(), testCase.generator())

			_, _, err := service.LoginUser(context.Background(), clientName, clientSecret, email, userPassword)
			require.Error(t, err)
		})
	}
}

func TestLogoutSuccess(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On("RevokeSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(int64(1), nil)

	service := session.NewService(mockStore, &user.MockService{}, &client.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), refreshToken)
	require.NoError(t, err)
}

func TestLogoutFailureWhenStoreCallFails(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On("RevokeSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(int64(0), liberr.WithArgs(errors.New("failed to revoke session")))

	service := session.NewService(mockStore, &user.MockService{}, &client.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), refreshToken)
	require.Error(t, err)
}

func TestRefreshTokenSuccess(t *testing.T) {
	ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
	require.NoError(t, err)

	mockStore := &session.MockStore{}
	mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(ss, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(newAccessToken, nil)

	mockClientService := &client.MockService{}
	mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 14440, nil)

	service := session.NewService(mockStore, &user.MockService{}, mockClientService, mockGenerator)

	_, err = service.RefreshToken(context.Background(), clientName, clientSecret, refreshToken)
	require.NoError(t, err)
}

func TestRefreshTokenFailure(t *testing.T) {
	testCases := map[string]struct {
		store         func() session.Store
		clientService func() client.Service
		generator     func() token.Generator
	}{
		"test failure when store call fails to get session": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(session.Session{}, liberr.WithArgs(errors.New("failed to get session")))

				return mockStore
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator:     func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when failed to get client ttl": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(session.Session{}, nil)

				return mockStore
			},
			clientService: func() client.Service {
				mockClientService := &client.MockService{}
				mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(-1, -1, liberr.WithArgs(errors.New("failed to get client ttl")))

				return mockClientService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when session is expired": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(ss, nil)
				mockStore.On("RevokeSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(int64(1), nil)

				return mockStore
			},
			clientService: func() client.Service {
				mockClientService := &client.MockService{}
				mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 87600, nil)

				return mockClientService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when failed to generate access token": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(ss, nil)

				return mockStore
			},
			clientService: func() client.Service {
				mockClientService := &client.MockService{}
				mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 87600, nil)

				return mockClientService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateAccessToken", 10, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("", liberr.WithArgs(errors.New("failed to generate token")))

				return mockGenerator
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := session.NewService(testCase.store(), &user.MockService{}, testCase.clientService(), testCase.generator())

			_, err := service.RefreshToken(context.Background(), clientName, clientSecret, refreshToken)
			require.Error(t, err)
		})
	}
}
