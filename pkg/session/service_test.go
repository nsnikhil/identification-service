package session_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
	"testing"
	"time"
)

type sessionTest struct {
	suite.Suite
	clientCfg         config.ClientConfig
	clientDefaultData map[string]interface{}
}

func (st *sessionTest) SetupSuite() {
	mockClientConfig := &config.MockClientConfig{}
	mockClientConfig.On("Strategies").
		Return(map[string]bool{test.ClientSessionStrategyRevokeOld: true})

	st.clientCfg = mockClientConfig
	st.clientDefaultData = map[string]interface{}{}
}

func TestClient(t *testing.T) {
	suite.Run(t, new(sessionTest))
}

func (st *sessionTest) TestLoginUserSuccess() {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	userEmail := test.NewEmail()
	sessionID := test.NewUUID()
	maxActiveSessions := test.RandInt(2, 10)
	accessTokenTTL := test.RandInt(1, 10)

	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(maxActiveSessions-1, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", accessTokenTTL, userID, map[string]string{"session_id": sessionID}).Return(test.NewPasetoToken(), nil)
	mockGenerator.On("GenerateRefreshToken").Return(test.NewUUID(), nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, mockUserService, mockGenerator, strategies)

	clientData := map[string]interface{}{
		test.ClientAccessTokenTTLKey:    accessTokenTTL,
		test.ClientMaxActiveSessionsKey: maxActiveSessions,
	}

	cl, err := test.NewClient(st.clientCfg, clientData)
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

	_, _, err = service.LoginUser(ctx, userEmail, userPassword)
	st.Require().NoError(err)
}

func (st *sessionTest) TestLoginUserSuccessWhenSessionCountExceed() {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	sessionID := test.NewUUID()
	userEmail := test.NewEmail()
	accessTokenTTL := test.RandInt(1, 10)

	mockStore := &session.MockStore{}
	mockStore.On("CreateSession", mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("Session")).Return(sessionID, nil)
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).Return(2, nil)
	mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), userID, 1).Return(int64(1), nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", accessTokenTTL, userID, map[string]string{"session_id": sessionID}).Return(test.NewPasetoToken(), nil)
	mockGenerator.On("GenerateRefreshToken").Return(test.NewUUID(), nil)

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, mockUserService, mockGenerator, strategies)

	clientData := map[string]interface{}{
		test.ClientAccessTokenTTLKey: accessTokenTTL,
	}

	cl, err := test.NewClient(st.clientCfg, clientData)
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

	_, _, err = service.LoginUser(ctx, userEmail, userPassword)
	st.Require().NoError(err)
}

func (st *sessionTest) TestLoginUserFailureWhenSessionCountExceed() {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	userEmail := test.NewEmail()
	maxActiveSession := test.RandInt(2, 10)

	mockStore := &session.MockStore{}
	mockStore.On("GetActiveSessionsCount", mock.AnythingOfType("*context.valueCtx"), userID).
		Return(maxActiveSession, nil)

	mockStore.On("RevokeLastNSessions", mock.AnythingOfType("*context.valueCtx"), userID, 1).
		Return(int64(0), errors.New("failed to revoke last n sessions"))

	mockUserService := &user.MockService{}
	mockUserService.On("GetUserID", mock.AnythingOfType("*context.valueCtx"), userEmail, userPassword).Return(userID, nil)

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, mockUserService, &token.MockGenerator{}, strategies)

	clientData := map[string]interface{}{
		test.ClientMaxActiveSessionsKey: maxActiveSession,
	}

	cl, err := test.NewClient(st.clientCfg, clientData)
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

	_, _, err = service.LoginUser(ctx, userEmail, userPassword)
	st.Require().Error(err)
}

func (st *sessionTest) TestLoginFailureWhenFailedToGetClientFromContext() {
	userPassword := test.NewPassword()

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(&session.MockStore{}),
	}

	service := session.NewService(&session.MockStore{}, &user.MockService{}, &token.MockGenerator{}, strategies)

	_, _, err := service.LoginUser(context.Background(), test.NewEmail(), userPassword)
	st.Require().Error(err)
}

func (st *sessionTest) TestLoginUserFailure() {
	userPassword := test.NewPassword()
	userID := test.NewUUID()
	userEmail := test.NewEmail()
	sessionID := test.NewUUID()
	maxActiveSessions := test.RandInt(1, 10)
	accessTokenTTL := test.RandInt(1, 10)

	clientData := map[string]interface{}{
		test.ClientAccessTokenTTLKey:    accessTokenTTL,
		test.ClientMaxActiveSessionsKey: maxActiveSessions,
	}

	cl, err := test.NewClient(st.clientCfg, clientData)
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

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
		st.Run(name, func() {
			strategies := map[string]session.Strategy{
				test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(testCase.store()),
			}

			service := session.NewService(testCase.store(), testCase.userService(), testCase.generator(), strategies)

			_, _, err := service.LoginUser(ctx, userEmail, userPassword)
			st.Require().Error(err)
		})
	}
}

func (st *sessionTest) TestLogoutSuccess() {
	refreshToken := test.NewUUID()

	ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
	st.Require().NoError(err)

	mockStore := &session.MockStore{}
	mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)
	mockStore.On(
		"RevokeSessions",
		mock.AnythingOfType("*context.valueCtx"),
		[]string{refreshToken},
	).Return(int64(1), nil)

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{}, strategies)

	cl, err := test.NewClient(st.clientCfg, map[string]interface{}{})
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

	err = service.LogoutUser(ctx, refreshToken)
	st.Require().NoError(err)
}

func (st *sessionTest) TestLogoutFailureWhenStoreCallFails() {
	refreshToken := test.NewUUID()
	cl, err := test.NewClient(st.clientCfg, map[string]interface{}{})
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

	testCases := map[string]struct {
		ctx   func() context.Context
		store func() session.Store
	}{
		"test failure when client is missing from the context": {
			ctx: func() context.Context {
				return context.Background()
			},
			store: func() session.Store {
				return &session.MockStore{}
			},
		},
		"test failure when get session call fails": {
			ctx: func() context.Context {
				return ctx
			},
			store: func() session.Store {
				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).
					Return(session.Session{}, errors.New("failed to fetch sessions"))
				return mockStore
			},
		},
		"test failure when validate session fails due to revoked session": {
			ctx: func() context.Context {
				return ctx
			},
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Revoked(true).Build()
				st.Require().NoError(err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).
					Return(ss, nil)
				return mockStore
			},
		},
		"test failure when validate session fails due to expired session": {
			ctx: func() context.Context {
				return ctx
			},
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().
					CreatedAt(time.Now().AddDate(0, -2, -1)).
					Build()

				st.Require().NoError(err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).
					Return(ss, nil)
				return mockStore
			},
		},
		"test failure when revoke session call fails": {
			ctx: func() context.Context {
				return ctx
			},
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
				st.Require().NoError(err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).
					Return(ss, nil)
				mockStore.On(
					"RevokeSessions",
					mock.AnythingOfType("*context.valueCtx"),
					[]string{refreshToken},
				).Return(int64(0), errors.New("failed to revoke session"))
				return mockStore
			},
		},
	}

	for name, testCase := range testCases {
		st.Run(name, func() {
			strategies := map[string]session.Strategy{
				test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(testCase.store()),
			}

			svc := session.NewService(testCase.store(), &user.MockService{}, &token.MockGenerator{}, strategies)

			err := svc.LogoutUser(testCase.ctx(), refreshToken)
			st.Assert().Error(err)
		})
	}

	//mockStore := &session.MockStore{}
	//mockStore.On(
	//	"RevokeSessions",
	//	mock.AnythingOfType("*context.emptyCtx"),
	//	[]string{refreshToken},
	//).Return(int64(0), errors.New("failed to revoke session"))
	//
	//
	//cl, err := test.NewClient(st.clientCfg, map[string]interface{}{})
	//st.Require().NoError(err)
	//
	//ctx, err := client.WithContext(context.Background(), cl)
	//st.Require().NoError(err)

}

func (st *sessionTest) TestRefreshTokenSuccess() {
	refreshToken := test.NewUUID()
	accessTokenTTL := test.RandInt(1, 10)

	ss, err := session.NewSessionBuilder().CreatedAt(time.Now()).Build()
	st.Require().NoError(err)

	mockStore := &session.MockStore{}
	mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)

	mockGenerator := &token.MockGenerator{}
	mockGenerator.On("GenerateAccessToken", accessTokenTTL, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]string")).Return(test.NewPasetoToken(), nil)

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, &user.MockService{}, mockGenerator, strategies)

	clientData := map[string]interface{}{
		test.ClientAccessTokenTTLKey: accessTokenTTL,
	}

	cl, err := test.NewClient(st.clientCfg, clientData)
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

	_, err = service.RefreshToken(ctx, refreshToken)
	st.Require().NoError(err)
}

func (st *sessionTest) TestRefreshTokenFailureWhenFailedToGetClientFromContext() {
	mockStore := &session.MockStore{}

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{}, strategies)

	_, err := service.RefreshToken(context.Background(), test.NewUUID())
	st.Require().Error(err)
}

func (st *sessionTest) TestRefreshTokenFailure() {
	refreshToken := test.NewUUID()
	accessTokenTTL := test.RandInt(1, 10)

	clientData := map[string]interface{}{
		test.ClientAccessTokenTTLKey: accessTokenTTL,
	}

	cl, err := test.NewClient(st.clientCfg, clientData)
	st.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	st.Require().NoError(err)

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
				st.Require().NoError(err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)
				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when session is revoked": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().
					CreatedAt(time.Now().AddDate(0, 1, 1)).
					Revoked(true).
					Build()
				st.Require().NoError(err)

				mockStore := &session.MockStore{}
				mockStore.On("GetSession", mock.AnythingOfType("*context.valueCtx"), refreshToken).Return(ss, nil)

				return mockStore
			},
			generator: func() token.Generator { return &token.MockGenerator{} },
		},
		"test failure when revoke session fails": {
			store: func() session.Store {
				ss, err := session.NewSessionBuilder().CreatedAt(time.Now().AddDate(0, -2, -1)).Build()
				st.Require().NoError(err)

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
				st.Require().NoError(err)

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
		st.Run(name, func() {
			strategies := map[string]session.Strategy{
				test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(testCase.store()),
			}

			service := session.NewService(testCase.store(), &user.MockService{}, testCase.generator(), strategies)

			_, err := service.RefreshToken(ctx, refreshToken)
			st.Require().Error(err)
		})
	}
}

func (st *sessionTest) TestRevokeAllSessionsSuccess() {
	userID := test.NewUUID()

	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"), userID,
	).Return(int64(1), nil)

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{}, strategies)

	err := service.RevokeAllSessions(context.Background(), userID)
	st.Require().NoError(err)
}

func (st *sessionTest) TestRevokeAllSessionsFailure() {
	userID := test.NewUUID()

	mockStore := &session.MockStore{}
	mockStore.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"), userID,
	).Return(int64(0), errors.New("failed to revoke all sessions"))

	strategies := map[string]session.Strategy{
		test.ClientSessionStrategyRevokeOld: session.NewRevokeOldStrategy(mockStore),
	}

	service := session.NewService(mockStore, &user.MockService{}, &token.MockGenerator{}, strategies)

	err := service.RevokeAllSessions(context.Background(), userID)
	st.Require().Error(err)
}
