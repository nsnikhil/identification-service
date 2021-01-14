package user_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/database"
	"identification-service/pkg/password"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"regexp"
	"testing"
)

type userStoreSuite struct {
	suite.Suite
	db    database.SQLDatabase
	mock  sqlmock.Sqlmock
	store user.Store
}

func (ust *userStoreSuite) SetupSuite() {
	sqlDB, mock := getMockDB(ust.T())

	ust.db = database.NewSQLDatabase(sqlDB, test.QueryTTL)
	ust.mock = mock

	ust.store = user.NewStore(ust.db)
}

func (ust *userStoreSuite) TestCreateUserSuccess() {
	name, email := test.RandString(8), test.NewEmail()

	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	query := `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, email, passwordHash, passwordSalt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(test.NewUUID()))

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	currUser, err := user.NewUserBuilder(mockEncoder).Name(name).Email(email).Password(userPassword).Build()
	require.NoError(ust.T(), err)

	_, err = ust.store.CreateUser(context.Background(), currUser)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestCreateUserFailure() {
	name, email := test.RandString(8), test.NewEmail()

	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	query := `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, email, passwordHash, passwordSalt).
		WillReturnError(errors.New("failed to create new User"))

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	currUser, err := user.NewUserBuilder(mockEncoder).Name(name).Email(email).Password(userPassword).Build()
	require.NoError(ust.T(), err)

	_, err = ust.store.CreateUser(context.Background(), currUser)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestGetUserSuccess() {
	userEmail := test.NewEmail()

	query := `select id, name, email, password_hash, password_salt from users where email = $1`

	rows := sqlmock.NewRows(
		[]string{
			"id", "name", "email", "passwordhash", "passwordsalt",
		},
	)

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userEmail).
		WillReturnRows(rows.AddRow("", "", "", "", ""))

	us := user.NewStore(ust.db)

	_, err := us.GetUser(context.Background(), userEmail)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestGetUserFailure() {
	userEmail := test.NewEmail()

	query := `select id, name, email, password_hash, password_salt from users where email = $1`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userEmail).
		WillReturnError(errors.New("failed to get data"))

	us := user.NewStore(ust.db)

	_, err := us.GetUser(context.Background(), userEmail)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestUpdatePasswordSuccess() {
	email := test.NewEmail()
	passwordSalt := test.RandBytes(86)
	passwordHash := test.RandString(44)

	query := `update users set password_hash=$1, password_salt=$2 where id=$3`

	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(passwordHash, passwordSalt, email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	us := user.NewStore(ust.db)

	_, err := us.UpdatePassword(context.Background(), email, passwordHash, passwordSalt)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestUpdatePasswordFailure() {
	email := test.NewEmail()
	passwordSalt := test.RandBytes(86)
	passwordHash := test.RandString(44)

	query := `update users set password_hash=$1, password_salt=$2 where id=$3`

	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(passwordHash, passwordSalt, email).
		WillReturnError(errors.New("failed to update password"))

	us := user.NewStore(ust.db)

	_, err := us.UpdatePassword(context.Background(), email, passwordHash, passwordSalt)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func TestStore(t *testing.T) {
	suite.Run(t, new(userStoreSuite))
}

func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return db, mock
}
