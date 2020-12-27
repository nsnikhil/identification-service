package client_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
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
	rd    *miniredis.Miniredis
	mock  sqlmock.Sqlmock
	store client.Store
}

func (cst *clientStoreSuite) SetupSuite() {
	rd, err := miniredis.Run()
	cst.Require().NoError(err)

	sqlDB, mock, err := sqlmock.New()
	cst.Require().NoError(err)

	cst.db = database.NewSQLDatabase(sqlDB, test.QueryTTL)
	cst.mock = mock
	cst.rd = rd

	cst.store = client.NewStore(cst.db, redis.NewClient(&redis.Options{Addr: rd.Addr()}))
}

func (cst *clientStoreSuite) TestCreateClientSuccess() {
	clientName, priKey := test.ClientName(), test.ClientPriKey()

	query := `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			clientName,
			test.ClientAccessTokenTTL,
			test.ClientSessionTTL,
			test.ClientMaxActiveSessions,
			test.ClientSessionStrategyRevokeOld,
			priKey,
		).WillReturnRows(sqlmock.NewRows([]string{"secret"}).AddRow(test.ClientSecret()))

	cl, err := client.NewClientBuilder().
		Name(clientName).
		AccessTokenTTL(test.ClientAccessTokenTTL).
		SessionTTL(test.ClientSessionTTL).
		MaxActiveSessions(test.ClientMaxActiveSessions).
		SessionStrategy(test.ClientSessionStrategyRevokeOld).
		PrivateKey(priKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestCreateClientFailure() {
	clientName, priKey := test.ClientName(), test.ClientPriKey()

	query := `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			clientName,
			test.ClientAccessTokenTTL,
			test.ClientSessionTTL,
			test.ClientMaxActiveSessions,
			test.ClientSessionStrategyRevokeOld,
			priKey,
		).WillReturnError(errors.New("failed to create client"))

	cl, err := client.NewClientBuilder().
		Name(clientName).
		AccessTokenTTL(test.ClientAccessTokenTTL).
		SessionTTL(test.ClientSessionTTL).
		MaxActiveSessions(test.ClientMaxActiveSessions).
		SessionStrategy(test.ClientSessionStrategyRevokeOld).
		PrivateKey(priKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientSuccess() {
	clientID := test.ClientID()

	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(clientID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := cst.store.RevokeClient(context.Background(), clientID)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientFailure() {
	clientID := test.ClientID()

	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(clientID).
		WillReturnError(errors.New("failed to revoke client"))

	_, err := cst.store.RevokeClient(context.Background(), clientID)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientSuccess() {
	name, secret := test.ClientName(), test.ClientSecret()

	query := `select id, revoked, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key from clients where name=$1 and secret=$2`

	rows := sqlmock.NewRows(
		[]string{"id", "revoked", "access_token_ttl", "session_ttl", "max_active_sessions", "session_strategy", "private_key"},
	).AddRow(
		test.ClientID(),
		false,
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
		test.ClientPriKey(),
	)

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(name, secret).
		WillReturnRows(rows)

	_, err := cst.store.GetClient(context.Background(), name, secret)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientFailure() {
	name, secret := test.ClientName(), test.ClientSecret()

	query := `select id, revoked, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key from clients where name=$1 and secret=$2`

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
