package session_test

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/password"
	"identification-service/pkg/queue"
	"identification-service/pkg/session"
	"identification-service/pkg/user"
	"testing"
)

type sessionStoreIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
	store       session.Store
	userService user.Service
}

func (sst *sessionStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	require.NoError(sst.T(), err)

	sst.db = db
	sst.store = session.NewStore(sst.db)
	encoder := password.NewEncoder(cfg.PasswordConfig())
	mockQueue := &queue.MockQueue{}
	mockQueue.On("UnsafePush", mock.AnythingOfType("[]uint8")).Return(nil)

	sst.userService = user.NewService(user.NewStore(db), encoder, mockQueue)
}

func (sst *sessionStoreIntegrationSuite) AfterTest(suiteName, testName string) {
	truncate(sst.T(), sst.db)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(context.Background(), newSession(sst.T(), userID, refreshToken))
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionFailureWhenUserIsNotPresent() {
	_, err := sst.store.CreateSession(context.Background(), newSession(sst.T(), userID, refreshToken))
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestCreateSessionFailureForDuplicateRefreshToken() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(context.Background(), newSession(sst.T(), userID, refreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.CreateSession(context.Background(), newSession(sst.T(), userID, refreshToken))
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetSessionSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(context.Background(), newSession(sst.T(), userID, refreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.GetSession(context.Background(), refreshToken)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestGetSessionFailure() {
	_, err := sst.store.GetSession(context.Background(), refreshToken)
	require.Error(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeSessionSuccess() {
	userID := createUser(sst)

	_, err := sst.store.CreateSession(context.Background(), newSession(sst.T(), userID, refreshToken))
	require.NoError(sst.T(), err)

	_, err = sst.store.RevokeSession(context.Background(), refreshToken)
	require.NoError(sst.T(), err)
}

func (sst *sessionStoreIntegrationSuite) TestRevokeSessionFailure() {
	_, err := sst.store.RevokeSession(context.Background(), refreshToken)
	require.Error(sst.T(), err)
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
	userID, err := sst.userService.CreateUser(context.Background(), name, email, userPassword)
	require.NoError(sst.T(), err)

	return userID
}

func TestSessionStoreIntegration(t *testing.T) {
	suite.Run(t, new(sessionStoreIntegrationSuite))
}

func truncate(t *testing.T, db *sql.DB) {
	_, err := db.Exec("truncate sessions")
	require.NoError(t, err)

	_, err = db.Exec("truncate users cascade ")
	require.NoError(t, err)
}
