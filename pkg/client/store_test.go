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
	"identification-service/pkg/config"
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
	cfg   config.ClientConfig
}

func (cst *clientStoreSuite) SetupSuite() {
	rd, err := miniredis.Run()
	cst.Require().NoError(err)

	sqlDB, mock, err := sqlmock.New()
	cst.Require().NoError(err)

	mockClientConfig := &config.MockClientConfig{}
	mockClientConfig.On("Strategies").
		Return(map[string]bool{test.ClientSessionStrategyRevokeOld: true})

	cst.db = database.NewSQLDatabase(sqlDB, test.QueryTTL)
	cst.mock = mock
	cst.rd = rd
	cst.cfg = mockClientConfig
	cst.store = client.NewStore(cst.db, redis.NewClient(&redis.Options{Addr: rd.Addr()}))
}

func (cst *clientStoreSuite) TestCreateClientSuccess() {
	accessTokenTTLVal := test.RandInt(1, 10)
	sessionTTLVal := test.RandInt(1440, 86701)
	maxActiveSessionsVal := test.RandInt(1, 10)
	clientName, priKey := test.RandString(8), test.ClientPriKey()

	query := `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			clientName,
			accessTokenTTLVal,
			sessionTTLVal,
			maxActiveSessionsVal,
			test.ClientSessionStrategyRevokeOld,
			priKey,
		).WillReturnRows(sqlmock.NewRows([]string{"secret"}).AddRow(test.NewUUID()))

	cl, err := client.NewClientBuilder(cst.cfg).
		Name(clientName).
		AccessTokenTTL(accessTokenTTLVal).
		SessionTTL(sessionTTLVal).
		MaxActiveSessions(maxActiveSessionsVal).
		SessionStrategy(test.ClientSessionStrategyRevokeOld).
		PrivateKey(priKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestCreateClientFailure() {
	accessTokenTTLVal := test.RandInt(1, 10)
	sessionTTLVal := test.RandInt(1440, 86701)
	maxActiveSessionsVal := test.RandInt(1, 10)

	clientName, priKey := test.RandString(8), test.ClientPriKey()

	query := `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`

	cst.mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			clientName,
			accessTokenTTLVal,
			sessionTTLVal,
			maxActiveSessionsVal,
			test.ClientSessionStrategyRevokeOld,
			priKey,
		).WillReturnError(errors.New("failed to create client"))

	cl, err := client.NewClientBuilder(cst.cfg).
		Name(clientName).
		AccessTokenTTL(accessTokenTTLVal).
		SessionTTL(sessionTTLVal).
		MaxActiveSessions(maxActiveSessionsVal).
		SessionStrategy(test.ClientSessionStrategyRevokeOld).
		PrivateKey(priKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(context.Background(), cl)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientSuccess() {
	clientID := test.NewUUID()

	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(clientID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := cst.store.RevokeClient(context.Background(), clientID)
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestRevokeClientFailure() {
	clientID := test.NewUUID()

	query := `update clients set revoked=true where id=$1`

	cst.mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(clientID).
		WillReturnError(errors.New("failed to revoke client"))

	_, err := cst.store.RevokeClient(context.Background(), clientID)
	require.Error(cst.T(), err)

	require.NoError(cst.T(), cst.mock.ExpectationsWereMet())
}

func (cst *clientStoreSuite) TestGetClientSuccess() {
	accessTokenTTLVal := test.RandInt(1, 10)
	sessionTTLVal := test.RandInt(1440, 86701)
	maxActiveSessionsVal := test.RandInt(1, 10)
	name, secret := test.RandString(8), test.NewUUID()

	query := `select id, revoked, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key from clients where name=$1 and secret=$2`

	rows := sqlmock.NewRows(
		[]string{"id", "revoked", "access_token_ttl", "session_ttl", "max_active_sessions", "session_strategy", "private_key"},
	).AddRow(
		test.NewUUID(),
		false,
		accessTokenTTLVal,
		sessionTTLVal,
		maxActiveSessionsVal,
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
	name, secret := test.RandString(8), test.NewUUID()

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
