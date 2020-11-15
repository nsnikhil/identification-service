package user_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/password"
	"identification-service/pkg/user"
	"regexp"
	"testing"
)

type userStoreSuite struct {
	suite.Suite
	db    *sql.DB
	mock  sqlmock.Sqlmock
	store user.Store
}

func (ust *userStoreSuite) SetupSuite() {
	ust.db, ust.mock = getMockDB(ust.T())
	ust.store = user.NewStore(ust.db)
}

func (ust *userStoreSuite) TestCreateUserSuccess() {
	query := `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, email, hash, salt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	currUser := getUser(ust.T(), name, email, userPassword)

	_, err := ust.store.CreateUser(context.Background(), currUser)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestCreateUserFailure() {
	query := `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, email, hash, salt).
		WillReturnError(errors.New("failed to create new User"))

	currUser := getUser(ust.T(), name, email, userPassword)

	_, err := ust.store.CreateUser(context.Background(), currUser)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestGetUserSuccess() {
	query := `select id, name, email, password_hash, password_salt from users where email = $1`

	rows := sqlmock.NewRows(
		[]string{
			"id", "name", "email", "passwordhash", "passwordsalt",
		},
	)

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(email).
		WillReturnRows(rows.AddRow("", "", "", "", ""))

	us := user.NewStore(ust.db)

	_, err := us.GetUser(context.Background(), email)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestGetUserFailure() {
	query := `select id, name, email, password_hash, password_salt from users where email = $1`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(email).
		WillReturnError(errors.New("failed to get data"))

	us := user.NewStore(ust.db)

	_, err := us.GetUser(context.Background(), email)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestUpdatePasswordSuccess() {
	query := `update users set password_hash=$1, password_salt=$2 where id=$3`

	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(hash, salt, email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	us := user.NewStore(ust.db)

	_, err := us.UpdatePassword(context.Background(), email, hash, salt)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestUpdatePasswordFailure() {
	query := `update users set password_hash=$1, password_salt=$2 where id=$3`

	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(hash, salt, email).
		WillReturnError(errors.New("failed to update password"))

	us := user.NewStore(ust.db)

	_, err := us.UpdatePassword(context.Background(), email, hash, salt)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func TestStore(t *testing.T) {
	suite.Run(t, new(userStoreSuite))
}

func getUser(t *testing.T, name, email, userPassword string) user.User {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(salt, nil)
	mockEncoder.On("GenerateKey", userPassword, salt).Return(key)
	mockEncoder.On("EncodeKey", key).Return(hash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	user, err := user.NewUser(mockEncoder, name, email, userPassword)
	require.NoError(t, err)

	return user
}

func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return db, mock
}
