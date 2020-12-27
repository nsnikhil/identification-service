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
	ust.db = test.NewDB(ust.T(), cfg)
	ust.ctx = context.Background()
	ust.store = user.NewStore(ust.db)
}

func (ust *userStoreIntegrationSuite) TearDownSuite() {
	truncate(ust)
}

func (ust *userStoreIntegrationSuite) TestCreateUserSuccess() {
	nu, _ := newUser(ust.T())

	_, err := ust.store.CreateUser(ust.ctx, nu)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestCreateUserFailureForDuplicateRecord() {
	nu, _ := newUser(ust.T())

	_, err := ust.store.CreateUser(ust.ctx, nu)
	require.NoError(ust.T(), err)

	_, err = ust.store.CreateUser(ust.ctx, nu)
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserSuccess() {
	nu, email := newUser(ust.T())

	_, err := ust.store.CreateUser(ust.ctx, nu)
	require.NoError(ust.T(), err)

	_, err = ust.store.GetUser(ust.ctx, email)
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestGetUserFailureWhenEmailIsNotPresent() {
	_, err := ust.store.GetUser(ust.ctx, test.UserEmail())
	require.Error(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordSuccessWithDB() {
	nu, _ := newUser(ust.T())

	id, err := ust.store.CreateUser(ust.ctx, nu)
	require.NoError(ust.T(), err)

	_, err = ust.store.UpdatePassword(ust.ctx, id, test.UserPasswordHash(), test.UserPasswordSalt())
	require.NoError(ust.T(), err)
}

func (ust *userStoreIntegrationSuite) TestUpdatePasswordFailureWhenUserIsNotPresent() {
	_, err := ust.store.UpdatePassword(
		ust.ctx,
		test.UserEmail(),
		test.UserPasswordHash(),
		test.UserPasswordSalt(),
	)

	require.Error(ust.T(), err)
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(userStoreIntegrationSuite))
}

func newUser(t *testing.T) (user.User, string) {
	userEmail := test.UserEmail()
	passwordSalt := test.UserPasswordSalt()
	passwordKey := test.UserPasswordKey()
	passwordHash := test.UserPasswordHash()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", test.UserPassword).Return(nil)

	us, err := user.NewUserBuilder(mockEncoder).
		Name(test.UserName()).
		Email(userEmail).
		Password(test.UserPassword).
		Build()

	require.NoError(t, err)

	return us, userEmail
}

func truncate(ust *userStoreIntegrationSuite) {
	_, err := ust.db.ExecContext(ust.ctx, "truncate users cascade ")
	require.NoError(ust.T(), err)
}
