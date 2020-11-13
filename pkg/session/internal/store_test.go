package internal_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/session/internal"
	"regexp"
	"testing"
	"time"
)

type sessionStoreSuite struct {
	suite.Suite
	mock  sqlmock.Sqlmock
	db    *sql.DB
	store internal.Store
}

func (st *sessionStoreSuite) SetupSuite() {
	st.db, st.mock = getMockDB(st.T())
	st.store = internal.NewStore(st.db)
}

func (st *sessionStoreSuite) TestCreateSessionSuccess() {
	query := `insert into sessions (userid, refreshtoken) values ($1, $2) returning id`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID, refreshToken).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(sessionID))

	s, err := internal.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	require.NoError(st.T(), err)

	_, err = st.store.CreateSession(context.Background(), s)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestCreateSessionFailure() {
	query := `insert into sessions (userid, refreshtoken) values ($1, $2) returning id`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(userID, refreshToken).
		WillReturnError(errors.New("failed to create session"))

	s, err := internal.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	require.NoError(st.T(), err)

	_, err = st.store.CreateSession(context.Background(), s)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetSessionSuccess() {
	query := `select id, userid, revoked, createdat, updatedat from sessions where refreshtoken=$1`

	rows := sqlmock.NewRows([]string{"id", "userid", "revoked", "createdat", "updatedat"}).
		AddRow(sessionID, userID, false, time.Time{}, time.Time{})

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(refreshToken).
		WillReturnRows(rows)

	_, err := st.store.GetSession(context.Background(), refreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestGetSessionFailure() {
	query := `select id, userid, revoked, createdat, updatedat from sessions where refreshtoken=$1`

	st.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(refreshToken).
		WillReturnError(errors.New("failed to get session"))

	_, err := st.store.GetSession(context.Background(), refreshToken)
	require.Error(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionSuccess() {
	query := `update sessions set revoked=true where refreshtoken=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(refreshToken).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := st.store.RevokeSession(context.Background(), refreshToken)
	require.NoError(st.T(), err)

	require.NoError(st.T(), st.mock.ExpectationsWereMet())
}

func (st *sessionStoreSuite) TestRevokeSessionFailure() {
	query := `update sessions set revoked=true where refreshtoken=$1`

	st.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(refreshToken).
		WillReturnError(errors.New("failed to revoke session"))

	_, err := st.store.RevokeSession(context.Background(), refreshToken)
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
