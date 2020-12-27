package user_test

//import (
//	"context"
//	"database/sql"
//	"errors"
//	"github.com/DATA-DOG/go-sqlmock"
//	"github.com/stretchr/testify/require"
//	"github.com/stretchr/testify/suite"
//	"identification-service/pkg/database"
//	"identification-service/pkg/password"
//	"identification-service/pkg/test"
//	"identification-service/pkg/user"
//	"regexp"
//	"testing"
//)
//
//type userStoreSuite struct {
//	suite.Suite
//	db    database.SQLDatabase
//	mock  sqlmock.Sqlmock
//	store user.Store
//}
//
//func (ust *userStoreSuite) SetupSuite() {
//	sqlDB, mock := getMockDB(ust.T())
//
//	ust.db = database.NewSQLDatabase(sqlDB, test.QueryTTL)
//	ust.mock = mock
//
//	ust.store = user.NewStore(ust.db)
//}
//
//func (ust *userStoreSuite) TestCreateUserSuccess() {
//	query := `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`
//
//	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
//		WithArgs(test.UserName, test.UserEmail, test.UserPasswordHash, test.UserPasswordSalt).
//		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(test.UserID))
//
//	currUser := getUser(ust.T(), test.UserName, test.UserEmail, test.UserPassword)
//
//	_, err := ust.store.CreateUser(context.Background(), currUser)
//	require.NoError(ust.T(), err)
//
//	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
//}
//
//func (ust *userStoreSuite) TestCreateUserFailure() {
//	query := `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`
//
//	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
//		WithArgs(test.UserName, test.UserEmail, test.UserPasswordHash, test.UserPasswordSalt).
//		WillReturnError(errors.New("failed to create new User"))
//
//	currUser := getUser(ust.T(), test.UserName, test.UserEmail, test.UserPassword)
//
//	_, err := ust.store.CreateUser(context.Background(), currUser)
//	require.Error(ust.T(), err)
//
//	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
//}
//
//func (ust *userStoreSuite) TestGetUserSuccess() {
//	query := `select id, name, email, password_hash, password_salt from users where email = $1`
//
//	rows := sqlmock.NewRows(
//		[]string{
//			"id", "name", "email", "passwordhash", "passwordsalt",
//		},
//	)
//
//	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
//		WithArgs(test.UserEmail).
//		WillReturnRows(rows.AddRow("", "", "", "", ""))
//
//	us := user.NewStore(ust.db)
//
//	_, err := us.GetUser(context.Background(), test.UserEmail)
//	require.NoError(ust.T(), err)
//
//	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
//}
//
//func (ust *userStoreSuite) TestGetUserFailure() {
//	query := `select id, name, email, password_hash, password_salt from users where email = $1`
//
//	ust.mock.ExpectQuery(regexp.QuoteMeta(query)).
//		WithArgs(test.UserEmail).
//		WillReturnError(errors.New("failed to get data"))
//
//	us := user.NewStore(ust.db)
//
//	_, err := us.GetUser(context.Background(), test.UserEmail)
//	require.Error(ust.T(), err)
//
//	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
//}
//
//func (ust *userStoreSuite) TestUpdatePasswordSuccess() {
//	query := `update users set password_hash=$1, password_salt=$2 where id=$3`
//
//	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
//		WithArgs(test.UserPasswordHash, test.UserPasswordSalt, test.UserEmail).
//		WillReturnResult(sqlmock.NewResult(1, 1))
//
//	us := user.NewStore(ust.db)
//
//	_, err := us.UpdatePassword(context.Background(), test.UserEmail, test.UserPasswordHash, test.UserPasswordSalt)
//	require.NoError(ust.T(), err)
//
//	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
//}
//
//func (ust *userStoreSuite) TestUpdatePasswordFailure() {
//	query := `update users set password_hash=$1, password_salt=$2 where id=$3`
//
//	ust.mock.ExpectExec(regexp.QuoteMeta(query)).
//		WithArgs(test.UserPasswordHash, test.UserPasswordSalt, test.UserEmail).
//		WillReturnError(errors.New("failed to update password"))
//
//	us := user.NewStore(ust.db)
//
//	_, err := us.UpdatePassword(context.Background(), test.UserEmail, test.UserPasswordHash, test.UserPasswordSalt)
//	require.Error(ust.T(), err)
//
//	require.NoError(ust.T(), ust.mock.ExpectationsWereMet())
//}
//
//func TestStore(t *testing.T) {
//	suite.Run(t, new(userStoreSuite))
//}
//
//func getUser(t *testing.T, name, email, userPassword string) user.User {
//	mockEncoder := &password.MockEncoder{}
//	mockEncoder.On("GenerateSalt").Return(test.UserPasswordSalt, nil)
//	mockEncoder.On("GenerateKey", userPassword, test.UserPasswordSalt).Return(test.UserPasswordKey)
//	mockEncoder.On("EncodeKey", test.UserPasswordKey).Return(test.UserPasswordHash)
//	mockEncoder.On("ValidatePassword", userPassword).Return(nil)
//
//	us, err := user.NewUserBuilder(mockEncoder).Name(name).Email(email).Password(userPassword).Build()
//	require.NoError(t, err)
//
//	return us
//}
//
//func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
//	db, mock, err := sqlmock.New()
//	require.NoError(t, err)
//
//	return db, mock
//}
