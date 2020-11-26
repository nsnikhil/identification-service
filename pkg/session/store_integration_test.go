package session_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/event/publisher"
	"identification-service/pkg/password"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"testing"
)

type sessionStoreIntegrationSuite struct {
	suite.Suite
	ctx         context.Context
	db          database.SQLDatabase
	store       session.Store
	userService user.Service
}

func (sst *sessionStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	dbCfg := cfg.DatabaseConfig()

	sqlDB, err := database.NewHandler(dbCfg).GetDB()
	require.NoError(sst.T(), err)

	db := database.NewSQLDatabase(sqlDB, dbCfg.QueryTTL())

	sst.ctx = context.Background()
	sst.db = db
	sst.store = session.NewStore(sst.db)
	encoder := password.NewEncoder(cfg.PasswordConfig())

	mockPublisher := &publisher.MockPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	sst.userService = user.NewService(user.NewStore(db), encoder, mockPublisher)
}

func (sst *sessionStoreIntegrationSuite) AfterTest(suiteName, testName string) {
	truncate(sst)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, test.SessionRefreshToken))
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionFailureWhenUserIsNotPresent() {
	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), test.UserID, test.SessionRefreshToken))
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionFailureForDuplicateRefreshToken() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, test.SessionRefreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, test.SessionRefreshToken))
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetSessionSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, test.SessionRefreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.GetSession(sst.ctx, test.SessionRefreshToken)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetSessionFailure() {
	_, err := sst.store.GetSession(sst.ctx, test.SessionRefreshToken)
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetActiveSessionsCountSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, test.SessionRefreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.GetActiveSessionsCount(sst.ctx, test.UserID)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeSessionsSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, test.SessionRefreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.RevokeSessions(sst.ctx, test.SessionRefreshToken)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeSessionsFailure() {
	_, err := sst.store.RevokeSessions(sst.ctx, test.SessionRefreshToken)
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeLastNSessionsSuccess() {
	userID := createUser(sst)

	rts := []string{
		test.SessionRefreshToken,
		test.SessionRefreshTokenTwo,
		test.SessionRefreshTokenThree,
		test.SessionRefreshTokenFour,
	}

	for _, rt := range rts {
		_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, rt))
		require.NoError(sst.T(), err)
	}

	c, err := sst.store.RevokeLastNSessions(sst.ctx, userID, 2)
	require.NoError(sst.T(), err)

	assert.Equal(sst.T(), int64(2), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeLastNSessionsFailure() {
	userID := createUser(sst)

	c, err := sst.store.RevokeLastNSessions(sst.ctx, userID, 2)
	require.Error(sst.T(), err)

	assert.Equal(sst.T(), int64(0), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeAllSessionsSuccess() {
	userID := createUser(sst)

	rts := []string{test.SessionRefreshToken, test.SessionRefreshTokenTwo}

	for _, rt := range rts {
		_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), userID, rt))
		require.NoError(sst.T(), err)
	}

	c, err := sst.store.RevokeAllSessions(sst.ctx, userID)
	require.NoError(sst.T(), err)

	assert.Equal(sst.T(), int64(2), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeAllSessionsFailureWhenNoSessionsExists() {
	userID := createUser(sst)

	c, err := sst.store.RevokeAllSessions(sst.ctx, userID)
	require.Error(sst.T(), err)

	assert.Equal(sst.T(), int64(0), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeAllSessionsFailureWhenUserDoesNotExists() {
	c, err := sst.store.RevokeAllSessions(sst.ctx, test.UserID)
	require.Error(sst.T(), err)

	assert.Equal(sst.T(), int64(0), c)
}

func newSession(t *testing.T, userID, refreshToken string) session.Session {
	ss, err := session.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	require.NoError(t, err)

	return ss
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(sessionStoreIntegrationSuite))
}

func createUser(sst *sessionStoreIntegrationSuite) string {
	userID, err := sst.userService.CreateUser(sst.ctx, test.UserName, test.UserEmail, test.UserPassword)
	require.NoError(sst.T(), err)

	return userID
}

func TestSessionStoreIntegration(t *testing.T) {
	suite.Run(t, new(sessionStoreIntegrationSuite))
}

func truncate(sst *sessionStoreIntegrationSuite) {
	_, err := sst.db.ExecContext(sst.ctx, "truncate sessions")
	require.NoError(sst.T(), err)

	_, err = sst.db.ExecContext(sst.ctx, "truncate users cascade ")
	require.NoError(sst.T(), err)
}
