package session_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"regexp"
	"testing"
	"time"
)

type sessionStoreSuite struct {
	suite.Suite
	mock  sqlmock.Sqlmock
	db    *sql.DB
	store session.Store
}

func (st *sessionStoreSuite) SetupSuite() {
	st.db, st.mock = getMockDB(st.T())
	st.store = session.NewStore(st.db)
}

func (st *sessionStoreSuite) TestCreateSessionSuccess() {
	query := `insert into sessions (user_id, refresh_token) values ($1, $2) returning id`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.UserID, test.SessionRefreshToken).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(test.SessionID))

	s, err := session.NewSessionBuilder().UserID(test.UserID).RefreshToken(test.SessionRefreshToken).Build()
	require.NoError(st.T(), err)

	_, err = st.store.CreateSession(context.Background(), s)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestCreateSessionFailure() {
	query := `insert into sessions (user_id, refresh_token) values ($1, $2) returning id`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.UserID, test.SessionRefreshToken).
		WillReturnError(errors.New("failed to create session"))

	s, err := session.NewSessionBuilder().UserID(test.UserID).RefreshToken(test.SessionRefreshToken).Build()
	require.NoError(st.T(), err)

	_, err = st.store.CreateSession(context.Background(), s)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetSessionSuccess() {
	query := `select id, user_id, revoked, created_at, updated_at from sessions where refresh_token=$1`

	rows := sqlmock.NewRows([]string{"id", "userid", "revoked", "createdat", "updatedat"}).
		AddRow(test.SessionID, test.UserID, false, time.Time{}, time.Time{})

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.SessionRefreshToken).
		WillReturnRows(rows)

	_, err := st.store.GetSession(context.Background(), test.SessionRefreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetSessionFailure() {
	query := `select id, user_id, revoked, created_at, updated_at from sessions where refresh_token=$1`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.SessionRefreshToken).
		WillReturnError(errors.New("failed to get session"))

	_, err := st.store.GetSession(context.Background(), test.SessionRefreshToken)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionSuccess() {
	query := `update sessions set revoked=true where refresh_token=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(test.SessionRefreshToken).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeSession(context.Background(), test.SessionRefreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionFailure() {
	query := `update sessions set revoked=true where refresh_token=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(test.SessionRefreshToken).
		WillReturnError(errors.New("failed to revoke session"))

	_, err := st.store.RevokeSession(context.Background(), test.SessionRefreshToken)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func TestSessionStore(t *testing.T) {
	suite.Run(t, new(sessionStoreSuite))
}

func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return db, mock
}
