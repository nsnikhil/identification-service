package internal_test

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/password"
	"identification-service/pkg/user/internal"
	"regexp"
	"testing"
)

type userStoreSuite struct {
	suite.Suite
	db    *sql.DB
	mock  sqlmock.Sqlmock
	store internal.Store
}

func (ust *userStoreSuite) SetupSuite() {
	ust.db, ust.mock = getMockDB(ust.T())
	ust.store = internal.NewStore(ust.db)
}

func (ust *userStoreSuite) TestCreateUserSuccess() {
	query := `insert into users (name, email, passwordhash, passwordsalt) values ($1, $2, $3, $4) returning id`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, email, hash, salt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

	currUser := getUser(ust.T(), name, email, userPassword)

	_, err := ust.store.CreateUser(currUser)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestCreateUserFailure() {
	query := `insert into users (name, email, passwordhash, passwordsalt) values ($1, $2, $3, $4) returning id`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, email, hash, salt).
		WillReturnError(errors.New("failed to create new User"))

	currUser := getUser(ust.T(), name, email, userPassword)

	_, err := ust.store.CreateUser(currUser)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestGetUserSuccess() {
	query := `select id, name, email, passwordhash, passwordsalt from users where email = $1`

	rows := sqlmock.NewRows(
		[]string{
			"id", "name", "email", "passwordhash", "passwordsalt",
		},
	)

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(email).
		WillReturnRows(rows.AddRow("", "", "", "", ""))

	us := internal.NewStore(ust.db)

	_, err := us.GetUser(email)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestGetUserFailure() {
	query := `select id, name, email, passwordhash, passwordsalt from users where email = $1`

	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(email).
		WillReturnError(errors.New("failed to get data"))

	us := internal.NewStore(ust.db)

	_, err := us.GetUser(email)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestUpdatePasswordSuccess() {
	query := `update users set passwordhash=$1, passwordsalt=$2 where id=$3`

	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(hash, salt, email).
		WillReturnResult(sqlmock.NewResult(1, 1))

	us := internal.NewStore(ust.db)

	_, err := us.UpdatePassword(email, hash, salt)
	require.NoError(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func (ust *userStoreSuite) TestUpdatePasswordFailure() {
	query := `update users set passwordhash=$1, passwordsalt=$2 where id=$3`

	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(hash, salt, email).
		WillReturnError(errors.New("failed to update password"))

	us := internal.NewStore(ust.db)

	_, err := us.UpdatePassword(email, hash, salt)
	require.Error(ust.T(), err)

	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
}

func TestStore(t *testing.T) {
	suite.Run(t, new(userStoreSuite))
}

func getUser(t *testing.T, name, email, userPassword string) internal.User {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(salt, nil)
	mockEncoder.On("GenerateKey", userPassword, salt).Return(key)
	mockEncoder.On("EncodeKey", key).Return(hash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	user, err := internal.NewUser(mockEncoder, name, email, userPassword)
	require.NoError(t, err)

	return user
}

func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return db, mock
}
