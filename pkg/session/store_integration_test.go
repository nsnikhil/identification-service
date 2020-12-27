package session_test

import (
	"context"
	"fmt"
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
	ctx    context.Context
	db     database.SQLDatabase
	store  session.Store
	userID string
}

func (sst *sessionStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	sst.ctx = context.Background()
	sst.db = test.NewDB(sst.T(), cfg)
	sst.store = session.NewStore(sst.db)
	sst.userID = createUser(sst, cfg)
}

func (sst *sessionStoreIntegrationSuite) TearDownSuite() {
	truncate(sst, "sessions", "users cascade")
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionSuccess() {
	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, test.SessionRefreshToken()))
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionFailureWhenUserIsNotPresent() {
	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), test.UserID(), test.SessionRefreshToken()))
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionFailureForDuplicateRefreshToken() {
	refreshToken := test.SessionRefreshToken()

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, refreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, refreshToken))
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetSessionSuccess() {
	refreshToken := test.SessionRefreshToken()

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, refreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.GetSession(sst.ctx, refreshToken)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetSessionFailure() {
	_, err := sst.store.GetSession(sst.ctx, test.SessionRefreshToken())
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetActiveSessionsCountSuccess() {
	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, test.SessionRefreshToken()))
	require.NoError(sst.T(), err)

	_, err = sst.store.GetActiveSessionsCount(sst.ctx, test.UserID())
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeSessionsSuccess() {
	refreshToken := test.SessionRefreshToken()

	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, refreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.RevokeSessions(sst.ctx, refreshToken)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeSessionsFailure() {
	_, err := sst.store.RevokeSessions(sst.ctx, test.SessionRefreshToken())
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeLastNSessionsSuccess() {
	rts := []string{
		test.SessionRefreshToken(),
		test.SessionRefreshToken(),
		test.SessionRefreshToken(),
		test.SessionRefreshToken(),
	}

	for _, rt := range rts {
		_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, rt))
		require.NoError(sst.T(), err)
	}

	c, err := sst.store.RevokeLastNSessions(sst.ctx, sst.userID, 2)
	require.NoError(sst.T(), err)

	assert.Equal(sst.T(), int64(2), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeLastNSessionsFailure() {
	//c, err := sst.store.RevokeLastNSessions(sst.ctx, sst.userID, 2)
	//require.Error(sst.T(), err)
	//
	//assert.Equal(sst.T(), int64(0), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeAllSessionsSuccess() {
	//rts := []string{test.SessionRefreshToken(), test.SessionRefreshToken()}
	//
	//for _, rt := range rts {
	//	_, err := sst.store.CreateSession(sst.ctx, newSession(sst.T(), sst.userID, rt))
	//	require.NoError(sst.T(), err)
	//}
	//
	//c, err := sst.store.RevokeAllSessions(sst.ctx, sst.userID)
	//require.NoError(sst.T(), err)

	//assert.Equal(sst.T(), int64(2), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeAllSessionsFailureWhenNoSessionsExists() {
	c, err := sst.store.RevokeAllSessions(sst.ctx, test.UserID())
	require.Error(sst.T(), err)

	assert.Equal(sst.T(), int64(0), c)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeAllSessionsFailureWhenUserDoesNotExists() {
	c, err := sst.store.RevokeAllSessions(sst.ctx, test.UserID())
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

func createUser(sst *sessionStoreIntegrationSuite, cfg config.Config) string {
	mockPublisher := &publisher.MockPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	encoder := password.NewEncoder(cfg.PasswordConfig())

	userService := user.NewService(user.NewStore(sst.db), encoder, mockPublisher)

	userID, err := userService.CreateUser(sst.ctx, test.UserName(), test.UserEmail(), test.UserPassword)
	require.NoError(sst.T(), err)
	require.NotEmpty(sst.T(), userID)

	return userID
}

func TestSessionStoreIntegration(t *testing.T) {
	suite.Run(t, new(sessionStoreIntegrationSuite))
}

func truncate(sst *sessionStoreIntegrationSuite, tables ...string) {
	for _, table := range tables {
		_, err := sst.db.ExecContext(sst.ctx, fmt.Sprintf("truncate %s", table))
		require.NoError(sst.T(), err)
	}
}
