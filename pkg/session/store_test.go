package session_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/database"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"regexp"
	"strings"
	"testing"
	"time"
)

type sessionStoreSuite struct {
	suite.Suite
	mock  sqlmock.Sqlmock
	db    database.SQLDatabase
	store session.Store
}

func (st *sessionStoreSuite) SetupSuite() {
	sqlDB, mock := getMockDB(st.T())

	st.db = database.NewSQLDatabase(sqlDB, test.QueryTTL)
	st.mock = mock
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

	rows := sqlmock.NewRows([]string{"id", "user_id", "revoked", "created_at", "updated_at"}).
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

func (st *sessionStoreSuite) TestGetActiveSessionsCountSuccess() {
	query := `select count(*) from sessions where user_id=$1 and revoked=false`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.UserID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := st.store.GetActiveSessionsCount(context.Background(), test.UserID)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetActiveSessionsCountFailure() {
	query := `select count(*) from sessions where user_id=$1 and revoked=false`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.UserID).
		WillReturnError(errors.New("failed to get active sessions count"))

	_, err := st.store.GetActiveSessionsCount(context.Background(), test.UserID)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionsSuccess() {
	query := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(toArgs([]string{test.SessionRefreshToken})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeSessions(context.Background(), test.SessionRefreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionsFailure() {
	query := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(toArgs([]string{test.SessionRefreshToken})).
		WillReturnError(errors.New("failed to revoke session"))

	_, err := st.store.RevokeSessions(context.Background(), test.SessionRefreshToken)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeLastNSessionsSuccess() {
	fetchQuery := `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`

	rows := sqlmock.NewRows([]string{"refresh_token"}).
		AddRow(test.SessionRefreshToken)

	st.mock.ExpectQuery(regexp.QuoteMeta(fetchQuery)).
		WithArgs(test.UserID, 1).
		WillReturnRows(rows)

	execQuery := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(execQuery)).
		WithArgs(toArgs([]string{test.SessionRefreshToken})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeLastNSessions(context.Background(), test.UserID, 1)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeLastNSessionsFailureWhenFetchFails() {
	fetchQuery := `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`

	st.mock.ExpectQuery(regexp.QuoteMeta(fetchQuery)).
		WithArgs(test.UserID, 1).
		WillReturnError(errors.New("failed to fetch refresh tokens"))

	_, err := st.store.RevokeLastNSessions(context.Background(), test.UserID, 1)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeLastNSessionsFailureWhenUpdateFails() {
	fetchQuery := `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`

	rows := sqlmock.NewRows([]string{"refresh_token"}).
		AddRow(test.SessionRefreshToken)

	st.mock.ExpectQuery(regexp.QuoteMeta(fetchQuery)).
		WithArgs(test.UserID, 1).
		WillReturnRows(rows)

	execQuery := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(execQuery)).
		WithArgs(toArgs([]string{test.SessionRefreshToken})).
		WillReturnError(errors.New("failed to revoke sessions"))

	_, err := st.store.RevokeLastNSessions(context.Background(), test.UserID, 1)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeAllSessionsSuccess() {
	query := `update sessions set revoked=true where user_id=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(test.UserID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeAllSessions(context.Background(), test.UserID)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeAllSessionsFailure() {
	query := `update sessions set revoked=true where user_id=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(test.UserID).
		WillReturnError(errors.New("failed to revoke all sessions"))

	_, err := st.store.RevokeAllSessions(context.Background(), test.UserID)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func toArgs(values []string) string {
	return "{" + strings.Join(values, ",") + "}"
}

func TestSessionStore(t *testing.T) {
	suite.Run(t, new(sessionStoreSuite))
}

func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return db, mock
}
