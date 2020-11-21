package client_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client"
	"identification-service/pkg/database"
	"identification-service/pkg/test"
	"regexp"
	"testing"
)

type clientStoreSuite struct {
	suite.Suite
	db    database.SQLDatabase
	mock  sqlmock.Sqlmock
	store client.Store
}

func (cst *clientStoreSuite) SetupSuite() {
	sqlDB, mock := getMockDB(cst.T())

	cst.db = database.NewSQLDatabase(sqlDB, test.QueryTTL)
	cst.mock = mock

	cst.store = client.NewStore(cst.db, &redis.Client{})
}

func (cst *clientStoreSuite) TestCreateClientSuccess() {
	query := `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			test.ClientName,
			test.ClientAccessTokenTTL,
			test.ClientSessionTTL,
			test.ClientMaxActiveSessions,
			test.ClientSessionStrategyRevokeOld,
			test.ClientPriKey,
		).WillReturnRows(sqlmock.NewRows([]string{"secret"}).AddRow(test.ClientSecret))

	cl, err := client.NewClientBuilder().
		Name(test.ClientName).
		AccessTokenTTL(test.ClientAccessTokenTTL).
		SessionTTL(test.ClientSessionTTL).
		MaxActiveSessions(test.ClientMaxActiveSessions).
		SessionStrategy(test.ClientSessionStrategyRevokeOld).
		PrivateKey(test.ClientPriKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestCreateClientFailure() {
	query := `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			test.ClientName,
			test.ClientAccessTokenTTL,
			test.ClientSessionTTL,
			test.ClientMaxActiveSessions,
			test.ClientSessionStrategyRevokeOld,
			test.ClientPriKey,
		).WillReturnError(errors.New("failed to create client"))

	cl, err := client.NewClientBuilder().
		Name(test.ClientName).
		AccessTokenTTL(test.ClientAccessTokenTTL).
		SessionTTL(test.ClientSessionTTL).
		MaxActiveSessions(test.ClientMaxActiveSessions).
		SessionStrategy(test.ClientSessionStrategyRevokeOld).
		PrivateKey(test.ClientPriKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientSuccess() {
	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(test.ClientID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := cst.store.RevokeClient(context.Background(), test.ClientID)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientFailure() {
	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(test.ClientID).
		WillReturnError(errors.New("failed to revoke client"))

	_, err := cst.store.RevokeClient(context.Background(), test.ClientID)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientSuccess() {
	query := `select id, revoked, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key from clients where name=$1 and secret=$2`

	rows := sqlmock.NewRows(
		[]string{"id", "revoked", "access_token_ttl", "session_ttl", "max_active_sessions", "session_strategy", "private_key"},
	).AddRow(
		test.ClientID, false,
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
		test.ClientPriKey,
	)

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.ClientName, test.ClientSecret).
		WillReturnRows(rows)

	_, err := cst.store.GetClient(context.Background(), test.ClientName, test.ClientSecret)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientFailure() {
	query := `select id, revoked, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key from clients where name=$1 and secret=$2`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(test.ClientName, test.ClientSecret).
		WillReturnError(errors.New("failed to get client"))

	_, err := cst.store.GetClient(context.Background(), test.ClientName, test.ClientSecret)
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
