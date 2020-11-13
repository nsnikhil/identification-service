package internal_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client/internal"
	"regexp"
	"testing"
)

const (
	id     = "f14abb31-ec1a-4ff6-a937-c2e930ca34ef"
	secret = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
)

type clientStoreSuite struct {
	suite.Suite
	db    *sql.DB
	mock  sqlmock.Sqlmock
	store internal.Store
}

func (cst *clientStoreSuite) SetupSuite() {
	cst.db, cst.mock = getMockDB(cst.T())
	cst.store = internal.NewStore(cst.db, &redis.Client{})
}

func (cst *clientStoreSuite) TestCreateClientSuccess() {
	query := `insert into clients (name, accesstokenttl, sessionttl) values ($1, $2, $3) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, accessTokenTTL, sessionTTL).
		WillReturnRows(sqlmock.NewRows([]string{"secret"}).AddRow(secret))

	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestCreateClientFailure() {
	query := `insert into clients (name, accesstokenttl, sessionttl) values ($1, $2, $3) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, accessTokenTTL, sessionTTL).
		WillReturnError(errors.New("failed to create client"))

	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientSuccess() {
	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := cst.store.RevokeClient(context.Background(), id)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientFailure() {
	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(id).
		WillReturnError(errors.New("failed to revoke client"))

	_, err := cst.store.RevokeClient(context.Background(), id)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientSuccess() {
	query := `select id, revoked, accesstokenttl, sessionttl from clients where name=$1 and secret=$2`

	rows := sqlmock.NewRows([]string{"id", "revoked", "accesstokenttl", "sessionttl"}).
		AddRow(id, false, accessTokenTTL, sessionTTL)

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, secret).
		WillReturnRows(rows)

	_, err := cst.store.GetClient(context.Background(), name, secret)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientFailure() {
	query := `select id, revoked, accesstokenttl, sessionttl from clients where name=$1 and secret=$2`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, secret).
		WillReturnError(errors.New("failed to get client"))

	_, err := cst.store.GetClient(context.Background(), name, secret)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func TestStore(t *testing.T) {
	suite.Run(t, new(clientStoreSuite))
}

func getMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	return db, mock
}
