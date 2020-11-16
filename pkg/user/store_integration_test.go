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
	"identification-service/pkg/test"
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

	_, err = ust.store.GetUser(context.Background(), test.UserEmail)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserFailureWhenEmailIsNotPresent() {
	_, err := ust.store.GetUser(context.Background(), test.UserEmail)
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordSuccessWithDB() {
	id, err := ust.store.CreateUser(context.Background(), newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.UpdatePassword(context.Background(), id, test.UserPasswordHash, test.UserPasswordSalt)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordFailureWhenUserIsNotPresent() {
	_, err := ust.store.UpdatePassword(
		context.Background(),
		test.UserEmail,
		test.UserPasswordHash,
		test.UserPasswordSalt,
	)

	require.Error(ust.T(), err)
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(userStoreIntegrationSuite))
}

func newUser(t *testing.T) user.User {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(test.UserPasswordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, test.UserPasswordSalt).Return(test.UserPasswordKey)
	mockEncoder.On("EncodeKey", test.UserPasswordKey).Return(test.UserPasswordHash)
	mockEncoder.On("ValidatePassword", test.UserPassword).Return(nil)

	us, err := user.NewUserBuilder(mockEncoder).
		Name(test.UserName).
		Email(test.UserEmail).
		Password(test.UserPassword).
		Build()

	require.NoError(t, err)

	return us
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
