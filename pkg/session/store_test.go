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
	userID, refreshToken := test.NewUUID(), test.NewUUID()

	query := `insert into sessions (user_id, refresh_token) values ($1, $2) returning id`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID, refreshToken).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(test.NewUUID()))

	s, err := session.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	require.NoError(st.T(), err)

	_, err = st.store.CreateSession(context.Background(), s)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestCreateSessionFailure() {
	userID, refreshToken := test.NewUUID(), test.NewUUID()

	query := `insert into sessions (user_id, refresh_token) values ($1, $2) returning id`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID, refreshToken).
		WillReturnError(errors.New("failed to create session"))

	s, err := session.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	require.NoError(st.T(), err)

	_, err = st.store.CreateSession(context.Background(), s)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetSessionSuccess() {
	refreshToken := test.NewUUID()

	query := `select id, user_id, revoked, created_at, updated_at from sessions where refresh_token=$1`

	rows := sqlmock.NewRows([]string{"id", "user_id", "revoked", "created_at", "updated_at"}).
		AddRow(test.NewUUID(), test.NewUUID(), false, time.Time{}, time.Time{})

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(refreshToken).
		WillReturnRows(rows)

	_, err := st.store.GetSession(context.Background(), refreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetSessionFailure() {
	refreshToken := test.NewUUID()

	query := `select id, user_id, revoked, created_at, updated_at from sessions where refresh_token=$1`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(refreshToken).
		WillReturnError(errors.New("failed to get session"))

	_, err := st.store.GetSession(context.Background(), refreshToken)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetActiveSessionsCountSuccess() {
	userID := test.NewUUID()

	query := `select count(*) from sessions where user_id=$1 and revoked=false`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := st.store.GetActiveSessionsCount(context.Background(), userID)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetActiveSessionsCountFailure() {
	userID := test.NewUUID()

	query := `select count(*) from sessions where user_id=$1 and revoked=false`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID).
		WillReturnError(errors.New("failed to get active sessions count"))

	_, err := st.store.GetActiveSessionsCount(context.Background(), userID)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionsSuccess() {
	refreshToken := test.NewUUID()

	query := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(toArgs([]string{refreshToken})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeSessions(context.Background(), refreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionsFailure() {
	refreshToken := test.NewUUID()

	query := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(toArgs([]string{refreshToken})).
		WillReturnError(errors.New("failed to revoke session"))

	_, err := st.store.RevokeSessions(context.Background(), refreshToken)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeLastNSessionsSuccess() {
	userID := test.NewUUID()
	refreshToken := test.NewUUID()

	fetchQuery := `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`

	rows := sqlmock.NewRows([]string{"refresh_token"}).
		AddRow(refreshToken)

	st.mock.ExpectQuery(regexp.QuoteMeta(fetchQuery)).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	execQuery := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(execQuery)).
		WithArgs(toArgs([]string{refreshToken})).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeLastNSessions(context.Background(), userID, 1)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeLastNSessionsFailureWhenFetchFails() {
	userID := test.NewUUID()

	fetchQuery := `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`

	st.mock.ExpectQuery(regexp.QuoteMeta(fetchQuery)).
		WithArgs(userID, 1).
		WillReturnError(errors.New("failed to fetch refresh tokens"))

	_, err := st.store.RevokeLastNSessions(context.Background(), userID, 1)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeLastNSessionsFailureWhenUpdateFails() {
	userID := test.NewUUID()
	refreshToken := test.NewUUID()

	fetchQuery := `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`

	rows := sqlmock.NewRows([]string{"refresh_token"}).
		AddRow(refreshToken)

	st.mock.ExpectQuery(regexp.QuoteMeta(fetchQuery)).
		WithArgs(userID, 1).
		WillReturnRows(rows)

	execQuery := `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`

	st.mock.ExpectExec(regexp.QuoteMeta(execQuery)).
		WithArgs(toArgs([]string{refreshToken})).
		WillReturnError(errors.New("failed to revoke sessions"))

	_, err := st.store.RevokeLastNSessions(context.Background(), userID, 1)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeAllSessionsSuccess() {
	userID := test.NewUUID()

	query := `update sessions set revoked=true where user_id=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeAllSessions(context.Background(), userID)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeAllSessionsFailure() {
	userID := test.NewUUID()

	query := `update sessions set revoked=true where user_id=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(userID).
		WillReturnError(errors.New("failed to revoke all sessions"))

	_, err := st.store.RevokeAllSessions(context.Background(), userID)
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
