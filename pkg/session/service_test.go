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
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	userEmail := test.NewEmail()
	sessionID := test.NewUUID()
	maxActiveSessions := test.RandInt(2, 10)
	accessTokenTTL := test.RandInt(1, 10)

	clientData := test.ClientData{AccessTokenTTL: accessTokenTTL, MaxActiveSession: maxActiveSessions}

	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(maxActiveSessions-1, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", accessTokenTTL, userID, map[string]string{"session_id": sessionID}).Return(test.NewPasetoToken(), nil)
	mockGenerator.On("GenerateRefreshToken").Return(test.NewUUID(), nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

	service := session.NewService(mockStore, mockUserService, mockGenerator)

	cl := test.NewClient(t, clientData)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, _, err = service.LoginUser(ctx, userEmail, userPassword)
	require.NoError(t, err)
}

func TestLoginUserSuccessWhenSessionCountExceed(t *testing.T) {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	sessionID := test.NewUUID()
	userEmail := test.NewEmail()
	accessTokenTTL := test.RandInt(1, 10)

	clientData := test.ClientData{AccessTokenTTL: accessTokenTTL}

	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(2, nil)
	mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), userID, 1).Return(int64(1), nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", accessTokenTTL, userID, map[string]string{"session_id": sessionID}).Return(test.NewPasetoToken(), nil)
	mockGenerator.On("GenerateRefreshToken").Return(test.NewUUID(), nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

	service := session.NewService(mockStore, mockUserService, mockGenerator)

	cl := test.NewClient(t, clientData)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, _, err = service.LoginUser(ctx, userEmail, userPassword)
	require.NoError(t, err)
}

func TestLoginUserFailureWhenSessionCountExceed(t *testing.T) {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	userEmail := test.NewEmail()
	maxActiveSession := test.RandInt(2, 10)

	clientData := test.ClientData{MaxActiveSession: maxActiveSession}

	mockStore := &session.MockStore{}
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).
		Return(maxActiveSession, nil)

	mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), userID, 1).
		Return(int64(0), errors.New("failed to revoke last n sessions"))

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

	service := session.NewService(mockStore, mockUserService, &token.MockGenerator{})

	cl := test.NewClient(t, clientData)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, _, err = service.LoginUser(ctx, userEmail, userPassword)
	require.Error(t, err)
}

func TestLoginFailureWhenFailedToGetClientFromContext(t *testing.T) {
	userPassword := test.NewPassword()
	service := session.NewService(&session.MockStore{}, &user.MockService{}, &token.MockGenerator{})

	_, _, err := service.LoginUser(context.Background(), test.NewEmail(), userPassword)
	require.Error(t, err)
}

func TestLoginUserFailure(t *testing.T) {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	userEmail := test.NewEmail()
	sessionID := test.NewUUID()
	maxActiveSessions := test.RandInt(1, 10)
	accessTokenTTL := test.RandInt(1, 10)

	clientData := test.ClientData{AccessTokenTTL: accessTokenTTL, MaxActiveSession: maxActiveSessions}

	ctx, err := client.WithContext(context.Background(), test.NewClient(t, clientData))
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
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return("", errors.New("failed to get user id"))

				return mockUserService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when get active session count fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(0, errors.New("failed to get active sessions count"))

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

				return mockUserService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when get active session count exceeds": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(maxActiveSessions+1, nil)
				mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), userID, 2).Return(int64(1), nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

				return mockUserService
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure get refresh token generation fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(maxActiveSessions-1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

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
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(maxActiveSessions-1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

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
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(test.NewUUID(), nil)

				return mockGenerator
			},
		},
		"test failure when access token generation fails": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)
				mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(1, nil)

				return mockStore
			},
			userService: func() user.Service {
				mockUserService := &user.MockService{}
				mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

				return mockUserService
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateRefreshToken").Return(test.NewUUID(), nil)
				mockGenerator.On("GenerateAccessToken", accessTokenTTL, userID, map[string]string{"session_id": sessionID}).Return("", errors.New("failed to generate access token"))

				return mockGenerator
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := session.NewService(testCase.store(), testCase.userService(), testCase.generator())

			_, _, err := service.LoginUser(ctx, userEmail, userPassword)
			require.Error(t, err)
		})
	}
}

func TestLogoutSuccess(t *testing.T) {
	refreshToken := test.NewUUID()

	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeSessions",
		mock.AnythingOfType("*context.emptyCtx"), []string{refreshToken},
	).Return(int64(1), nil)

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), refreshToken)
	require.NoError(t, err)
}

func TestLogoutFailureWhenStoreCallFails(t *testing.T) {
	refreshToken := test.NewUUID()

	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeSessions",
		mock.AnythingOfType("*context.emptyCtx"), []string{refreshToken},
	).Return(int64(0), errors.New("failed to revoke session"))

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.LogoutUser(context.Background(), refreshToken)
	require.Error(t, err)
}

func TestRefreshTokenSuccess(t *testing.T) {
	refreshToken := test.NewUUID()
	accessTokenTTL := test.RandInt(1, 10)

	ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
	require.NoError(t, err)

	mockStore := &session.MockStore{}
	mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", accessTokenTTL, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(test.NewPasetoToken(), nil)

	service := session.NewService(mockStore, &user.MockService{}, mockGenerator)

	cl := test.NewClient(t, test.ClientData{AccessTokenTTL: accessTokenTTL})

	require.NoError(t, err)

	ctx, err := client.WithContext(context.Background(), cl)
	require.NoError(t, err)

	_, err = service.RefreshToken(ctx, refreshToken)
	require.NoError(t, err)
}

func TestRefreshTokenFailureWhenFailedToGetClientFromContext(t *testing.T) {
	service := session.NewService(&session.MockStore{}, &user.MockService{}, &token.MockGenerator{})

	_, err := service.RefreshToken(context.Background(), test.NewUUID())
	require.Error(t, err)
}

func TestRefreshTokenFailure(t *testing.T) {
	refreshToken := test.NewUUID()
	accessTokenTTL := test.RandInt(1, 10)

	ctx, err := client.WithContext(context.Background(), test.NewClient(t, test.ClientData{AccessTokenTTL: accessTokenTTL}))
	require.NoError(t, err)

	testCases := map[string]struct {
		store     func() session.Store
		generator func() token.Generator
	}{
		"test failure when store call fails to get session": {
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(session.Session{}, errors.New("failed to get session"))

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when session is expired": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)
				mockStore.On("RevokeSessions", mock.AnythingOfType("*context.valueCtx"), []string{refreshToken}).Return(int64(1), nil)

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when revoke session fails": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)
				mockStore.On("RevokeSessions", mock.AnythingOfType("*context.valueCtx"), []string{refreshToken}).Return(int64(0), errors.New("failed to revoke session"))

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when failed to generate access token": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
				require.NoError(t, err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)

				return mockStore
			},
			generator: func() token.Generator {
				mockGenerator := &token.MockGenerator{}
				mockGenerator.On("GenerateAccessToken", accessTokenTTL, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return("", errors.New("failed to generate token"))

				return mockGenerator
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := session.NewService(testCase.store(), &user.MockService{}, testCase.generator())

			_, err := service.RefreshToken(ctx, refreshToken)
			require.Error(t, err)
		})
	}
}

func TestRevokeAllSessionsSuccess(t *testing.T) {
	userID := test.NewUUID()

	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"), userID,
	).Return(int64(1), nil)

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.RevokeAllSessions(context.Background(), userID)
	require.NoError(t, err)
}

func TestRevokeAllSessionsFailure(t *testing.T) {
	userID := test.NewUUID()

	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"), userID,
	).Return(int64(0), errors.New("failed to revoke all sessions"))

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{})

	err := service.RevokeAllSessions(context.Background(), userID)
	require.Error(t, err)
}
