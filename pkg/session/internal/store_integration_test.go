package internal_test

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
	"identification-service/pkg/session/internal"
	"identification-service/pkg/user"
	"testing"
)

var salt = []byte{90, 20, 247, 194, 220, 48, 153, 58, 158, 103, 9, 17, 243, 24, 179, 254, 88, 59, 161, 81, 216, 8, 126, 122, 102, 151, 200, 12, 134, 118, 146, 197, 193, 248, 117, 57, 127, 137, 112, 233, 116, 50, 128, 84, 127, 93, 180, 23, 81, 69, 245, 183, 45, 57, 51, 125, 9, 46, 200, 175, 97, 49, 11, 0, 40, 228, 186, 60, 177, 43, 69, 52, 168, 195, 69, 101, 21, 245, 62, 131, 252, 96, 240, 154, 251, 2}
var key = []byte{34, 179, 107, 154, 0, 94, 48, 1, 134, 44, 128, 127, 254, 17, 124, 248, 69, 96, 196, 174, 146, 255, 131, 91, 94, 143, 105, 33, 230, 157, 77, 243}
var hash = "IrNrmgBeMAGGLIB//hF8+EVgxK6S/4NbXo9pIeadTfM="

type sessionStoreIntegrationSuite struct {
	suite.Suite
	db          *sql.DB
	store       internal.Store
	userService user.Service
}

func (sst *sessionStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../../local.env")

	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	require.NoError(sst.T(), err)

	sst.db = db
	sst.store = internal.NewStore(sst.db)
	encoder := password.NewEncoder(cfg.PasswordConfig())
	mockQueue := &queue.MockQueue{}
	mockQueue.On("UnsafePush", mock.AnythingOfType("[]uint8")).Return(nil)

	sst.userService = user.NewService(sst.db, encoder, mockQueue)
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

func newSession(t *testing.T, userID, refreshToken string) internal.Session {
	ss, err := internal.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
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
