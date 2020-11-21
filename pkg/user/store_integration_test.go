// build integration_test

package user_test

import (
	"context"
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
	db    database.SQLDatabase
	ctx   context.Context
	store user.Store
}

func (ust *userStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	dbCfg := cfg.DatabaseConfig()

	sqlDB, err := database.NewHandler(dbCfg).GetDB()
	require.NoError(ust.T(), err)

	db := database.NewSQLDatabase(sqlDB, dbCfg.QueryTTL())

	ust.db = db
	ust.ctx = context.Background()
	ust.store = user.NewStore(ust.db)
}

func (ust *userStoreIntegrationSuite) AfterTest(suiteName, testName string) {
	truncate(ust)
}

func (ust *userStoreIntegrationSuite) TestCreateUserSuccess() {
	_, err := ust.store.CreateUser(ust.ctx, newUser(ust.T()))
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestCreateUserFailureForDuplicateRecord() {
	_, err := ust.store.CreateUser(ust.ctx, newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.CreateUser(ust.ctx, newUser(ust.T()))
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserSuccess() {
	_, err := ust.store.CreateUser(ust.ctx, newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.GetUser(ust.ctx, test.UserEmail)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserFailureWhenEmailIsNotPresent() {
	_, err := ust.store.GetUser(ust.ctx, test.UserEmail)
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordSuccessWithDB() {
	id, err := ust.store.CreateUser(ust.ctx, newUser(ust.T()))
	require.NoError(ust.T(), err)

	_, err = ust.store.UpdatePassword(ust.ctx, id, test.UserPasswordHash, test.UserPasswordSalt)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordFailureWhenUserIsNotPresent() {
	_, err := ust.store.UpdatePassword(
		ust.ctx,
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

func truncate(ust *userStoreIntegrationSuite) {
	_, err := ust.db.ExecContext(ust.ctx, "truncate users cascade ")
	require.NoError(ust.T(), err)
}
