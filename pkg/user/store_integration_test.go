// build integration_test

package user_test

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/password"
	"identification-service/pkg/user"
	"testing"
)

type userStoreIntegrationSuite struct {
	suite.Suite
	db    *sql.DB
	store user.Store
}

func (ust *userStoreIntegrationSuite) SetupSuite() {
	ust.db = getDB(ust.T())
	truncate(ust.T(), ust.db)
	ust.store = user.NewStore(ust.db)
}

func (ust *userStoreIntegrationSuite) AfterTest(suiteName, testName string) {
	truncate(ust.T(), ust.db)
}

func (ust *userStoreIntegrationSuite) TestCreateUserSuccess() {
	_, err := ust.store.CreateUser(context.Background(), newUser(ust.T()))
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestCreateUserFailureForDuplicateRecord() {
	_, err := ust.store.CreateUser(context.Background(), newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.CreateUser(context.Background(), newUser(ust.T()))
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserSuccess() {
	_, err := ust.store.CreateUser(context.Background(), newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.GetUser(context.Background(), email)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserFailureWhenEmailIsNotPresent() {
	_, err := ust.store.GetUser(context.Background(), email)
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordSuccessWithDB() {
	id, err := ust.store.CreateUser(context.Background(), newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.UpdatePassword(context.Background(), id, hash, salt)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordFailureWhenUserIsNotPresent() {
	_, err := ust.store.UpdatePassword(context.Background(), email, hash, salt)
	require.Error(ust.T(), err)
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(userStoreIntegrationSuite))
}

func newUser(t *testing.T) user.User {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(salt, nil)
	mockEncoder.On("GenerateKey", userPassword, salt).Return(key)
	mockEncoder.On("EncodeKey", key).Return(hash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	user, err := user.NewUser(mockEncoder, name, email, userPassword)
	require.NoError(t, err)

	return user
}

func truncate(t *testing.T, db *sql.DB) {
	_, err := db.Exec("truncate users cascade ")
	require.NoError(t, err)
}

func getDB(t *testing.T) *sql.DB {
	cfg := config.NewConfig("../../local.env")

	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	require.NoError(t, err)

	return db
}
