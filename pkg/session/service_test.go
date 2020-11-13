package session_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/liberr"
	"identification-service/pkg/session"
	"identification-service/pkg/session/internal"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
	"testing"
	"time"
)

const (
	email        = "test@test.com"
	userPassword = "Password@1234"

	clientName   = "clientOne"
	clientSecret = "86d690dd-92a0-40ac-ad48-110c951e3cb8"

	userID    = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
	sessionID = "f113fe5c-de2f-4876-b734-b51fbdc96e4b"

	accessToken  = "v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMDozNjowNyswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTA6MjY6MDcrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiMTEwMTI0NjUtMDNhNC00OWI2LTgwODEtY2RmYzczMDlhY2MwIiwibmJmIjoiMjAyMC0xMS0wN1QxMDoyNjowNyswNTozMCJ9PrXViH5779NxXHK_PxnwW-FdFV0klU07umd8X7F0A9irFLX7GTS3AczNm_hmb_yfYOX0o4DJri89AWeCb0qTAg.bnVsbA"
	refreshToken = "5df8159e-fd51-4e6c-9849-a9b1f070a403"

	newAccessToken = "v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMjozNDowOCswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTI6MjQ6MDgrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiZjJiNzhlNWYtNTZhMi00MzMwLWFhYWUtYmM4OWM1NzllNzIwIiwibmJmIjoiMjAyMC0xMS0wN1QxMjoyNDowOCswNTozMCIsInN1YiI6Ijg2ZDY5MGRkLTkyYTAtNDBhYy1hZDQ4LTExMGM5NTFlM2NiOCJ9DHCzvrlz6_QDB6zuuQcAmZs6yFoqBgkcHbtIVRcsDJ068XGs6N5R4U069lQvy-r7fHY2pL6tmxjRAZq1McetAA.bnVsbA"
)

func TestLoginUserSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.Session")).Return(sessionID, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, userID, map[string]string{"session_id": "f113fe5c-de2f-4876-b734-b51fbdc96e4b"}).Return(accessToken, nil)
	mockGenerator.On("GenerateRefreshToken").Return(refreshToken, nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return(userID, nil)

	mockClientService := &client.MockService{}
	mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 14440, nil)

	service := session.NewInternalService(mockStore, mockUserService, mockClientService, mockGenerator)

	_, _, err := service.LoginUser(context.Background(), clientName, clientSecret, email, userPassword)
	require.NoError(t, err)
}

func TestLoginUserFailure(t *testing.T) {
	testCases := map[string]struct {
		store         func() internal.Store
		userService   func() user.Service
		clientService func() client.Service
		generator     func() token.Generator
	}{
		"test failure get user id fails": {
			store: func() internal.Store { return &internal.MockStore{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.emptyCtx"), email, userPassword).Return("", liberr.WithArgs(errors.New("failed to get user id")))

				return mockUserService
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator:     func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure get refresh token generation fails": {
			store: func() internal.Store { return &internal.MockStore{} },
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
			store: func() internal.Store { return &internal.MockStore{} },
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
			store: func() internal.Store { return &internal.MockStore{} },
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
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.Session")).Return("", liberr.WithArgs(errors.New("failed to create new session")))

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
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.Session")).Return(sessionID, nil)

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
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.Session")).Return(sessionID, nil)

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
			service := session.NewInternalService(testCase.store(), testCase.userService(), testCase.clientService(), testCase.generator())

			_, _, err := service.LoginUser(context.Background(), clientName, clientSecret, email, userPassword)
			require.Error(t, err)
		})
	}
}

func TestLogoutSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("RevokeSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(int64(1), nil)

	service := session.NewInternalService(mockStore, &user.MockService{}, &client.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), refreshToken)
	require.NoError(t, err)
}

func TestLogoutFailureWhenStoreCallFails(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("RevokeSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(int64(0), liberr.WithArgs(errors.New("failed to revoke session")))

	service := session.NewInternalService(mockStore, &user.MockService{}, &client.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), refreshToken)
	require.Error(t, err)
}

func TestRefreshTokenSuccess(t *testing.T) {
	ss, err := internal.NewSessionBuilder().CreatedAt(time.Now()).Build()
	require.NoError(t, err)

	mockStore := &internal.MockStore{}
	mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(ss, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(newAccessToken, nil)

	mockClientService := &client.MockService{}
	mockClientService.On("GetClientTTL", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(10, 14440, nil)

	service := session.NewInternalService(mockStore, &user.MockService{}, mockClientService, mockGenerator)

	_, err = service.RefreshToken(context.Background(), clientName, clientSecret, refreshToken)
	require.NoError(t, err)
}

func TestRefreshTokenFailure(t *testing.T) {
	testCases := map[string]struct {
		store         func() internal.Store
		clientService func() client.Service
		generator     func() token.Generator
	}{
		"test failure when store call fails to get session": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(internal.Session{}, liberr.WithArgs(errors.New("failed to get session")))

				return mockStore
			},
			clientService: func() client.Service { return &client.MockService{} },
			generator:     func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when failed to get client ttl": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.timerCtx"), refreshToken).Return(internal.Session{}, nil)

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
			store: func() internal.Store {
				ss, err := internal.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				require.NoError(t, err)

				mockStore := &internal.MockStore{}
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
			store: func() internal.Store {
				ss, err := internal.NewSessionBuilder().CreatedAt(time.Now()).Build()
				require.NoError(t, err)

				mockStore := &internal.MockStore{}
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
			service := session.NewInternalService(testCase.store(), &user.MockService{}, testCase.clientService(), testCase.generator())

			_, err := service.RefreshToken(context.Background(), clientName, clientSecret, refreshToken)
			require.Error(t, err)
		})
	}
}
