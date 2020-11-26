package session_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
	"testing"
	"time"
)

//TODO: FIX LINE BREAK ON THIS FILE
func TestLoginUserSuccess(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(test.SessionID, nil)
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(1, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, test.UserID, map[string]string{"session_id": "f113fe5c-de2f-4876-b734-b51fbdc96e4b"}).Return(test.SessionAccessToken, nil)
	mockGenerator.On("GenerateRefreshToken").Return(test.SessionRefreshToken, nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

	service := session.NewService(mockStore, mockUserService, mockGenerator)

	cl := test.NewClient(t)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, _, err = service.LoginUser(ctx, test.UserEmail, test.UserPassword)
	require.NoError(t, err)
}

func TestLoginUserSuccessWhenSessionCountExceed(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(test.SessionID, nil)
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(2, nil)
	mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), test.UserID, 1).Return(int64(1), nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, test.UserID, map[string]string{"session_id": "f113fe5c-de2f-4876-b734-b51fbdc96e4b"}).Return(test.SessionAccessToken, nil)
	mockGenerator.On("GenerateRefreshToken").Return(test.SessionRefreshToken, nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

	service := session.NewService(mockStore, mockUserService, mockGenerator)

	cl := test.NewClient(t)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, _, err = service.LoginUser(ctx, test.UserEmail, test.UserPassword)
	require.NoError(t, err)
}

func TestLoginUserFailureWhenSessionCountExceed(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(2, nil)
	mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), test.UserID, 1).Return(int64(0), errors.New("failed to revoke last n sessions"))

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

	service := session.NewService(mockStore, mockUserService, &token.MockGenerator{})

	cl := test.NewClient(t)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, _, err = service.LoginUser(ctx, test.UserEmail, test.UserPassword)
	require.Error(t, err)
}

func TestLoginFailureWhenFailedToGetClientFromContext(t *testing.T) {
	service := session.NewService(&session.MockStore{}, &user.MockService{}, &token.MockGenerator{})

	_, _, err := service.LoginUser(context.Background(), test.UserEmail, test.UserPassword)
	require.Error(t, err)
}

func TestLoginUserFailure(t *testing.T) {
	ctx, err := client.WithContext(context.Background(), test.NewClient(t))
	require.NoError(t, err)

	testCases := map[string]struct {
		store       func() session.Store
		userService func() user.Service
		generator   func() token.Generator
	}{
		"test failure get user id fails": {
			store: func() session.Store { return &session.MockStore{} },
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return("", errors.New("failed to get user id"))

				return mockUserService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when get active session count fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(0, errors.New("failed to get active sessions count"))

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

				return mockUserService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when get active session count exceeds": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(3, nil)
				mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), test.UserID, 2).Return(int64(1), nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

				return mockUserService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure get refresh token generation fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return("", errors.New("failed to generate refresh token"))

				return mockGenerator
			},
		},
		"test failure when session creation fails due to invalid refresh token": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return("invalidRefreshToken", nil)

				return mockGenerator
			},
		},
		"test failure when store call fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return("", errors.New("failed to create new session"))
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(test.SessionRefreshToken, nil)

				return mockGenerator
			},
		},
		"test failure when access token generation fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(test.SessionID, nil)
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), test.UserID).Return(1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), test.UserEmail, test.UserPassword).Return(test.UserID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(test.SessionRefreshToken, nil)
				mockGenerator.On("GenerateAccessToken", 10, test.UserID, map[string]string{"session_id": "f113fe5c-de2f-4876-b734-b51fbdc96e4b"}).Return("", errors.New("failed to generate access token"))

				return mockGenerator
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := session.NewService(testCase.store(), testCase.userService(), testCase.generator())

			_, _, err := service.LoginUser(ctx, test.UserEmail, test.UserPassword)
			require.Error(t, err)
		})
	}
}

func TestLogoutSuccess(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeSessions",
		mock.AnythingOfType("*context.emptyCtx"), []string{test.SessionRefreshToken},
	).Return(int64(1), nil)

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), test.SessionRefreshToken)
	require.NoError(t, err)
}

func TestLogoutFailureWhenStoreCallFails(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeSessions",
		mock.AnythingOfType("*context.emptyCtx"), []string{test.SessionRefreshToken},
	).Return(int64(0), errors.New("failed to revoke session"))

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), test.SessionRefreshToken)
	require.Error(t, err)
}

func TestRefreshTokenSuccess(t *testing.T) {
	ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
	require.NoError(t, err)

	mockStore := &session.MockStore{}
	mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), test.SessionRefreshToken).Return(ss, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", 10, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(test.SessionAccessTokenTwo, nil)

	service := session.NewService(mockStore, &user.MockService{}, mockGenerator)

	cl := test.NewClient(t)

	require.NoError(t, err)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, err = service.RefreshToken(ctx, test.SessionRefreshToken)
	require.NoError(t, err)
}

func TestRefreshTokenFailureWhenFailedToGetClientFromContext(t *testing.T) {
	service := session.NewService(&session.MockStore{}, &user.MockService{}, &token.MockGenerator{})

	_, err := service.RefreshToken(context.Background(), test.SessionRefreshToken)
	require.Error(t, err)
}

func TestRefreshTokenFailure(t *testing.T) {
	ctx, err := client.WithContext(context.Background(), test.NewClient(t))
	require.NoError(t, err)

	testCases := map[string]struct {
		store     func() session.Store
		generator func() token.Generator
	}{
		"test failure when store call fails to get session": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), test.SessionRefreshToken).Return(session.Session{}, errors.New("failed to get session"))

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when session is expired": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), test.SessionRefreshToken).Return(ss, nil)
				mockStore.On("RevokeSessions", mock.AnythingOfType("*context.valueCtx"), []string{test.SessionRefreshToken}).Return(int64(1), nil)

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when revoke session fails": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), test.SessionRefreshToken).Return(ss, nil)
				mockStore.On("RevokeSessions", mock.AnythingOfType("*context.valueCtx"), []string{test.SessionRefreshToken}).Return(int64(0), errors.New("failed to revoke session"))

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when failed to generate access token": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), test.SessionRefreshToken).Return(ss, nil)

				return mockStore
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateAccessToken", 10, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("", errors.New("failed to generate token"))

				return mockGenerator
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := session.NewService(testCase.store(), &user.MockService{}, testCase.generator())

			_, err := service.RefreshToken(ctx, test.SessionRefreshToken)
			require.Error(t, err)
		})
	}
}

func TestRevokeAllSessionsSuccess(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"), test.UserID,
	).Return(int64(1), nil)

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.RevokeAllSessions(context.Background(), test.UserID)
	require.NoError(t, err)
}

func TestRevokeAllSessionsFailure(t *testing.T) {
	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"), test.UserID,
	).Return(int64(0), errors.New("failed to revoke all sessions"))

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.RevokeAllSessions(context.Background(), test.UserID)
	require.Error(t, err)
}
