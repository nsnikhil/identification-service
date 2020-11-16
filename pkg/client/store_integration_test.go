// build integration_test

package client_test

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/cache"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"testing"
)

type clientStoreIntegrationSuite struct {
	suite.Suite
	db    *sql.DB
	cache *redis.Client
	store client.Store
	ctx   context.Context
}

func (cst *clientStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	require.NoError(cst.T(), err)

	cc, err := cache.NewHandler(cfg.CacheConfig()).GetCache()
	require.NoError(cst.T(), err)

	cst.db = db
	cst.cache = cc
	cst.store = client.NewStore(cst.db, cst.cache)
	cst.ctx = context.Background()
}

func (cst *clientStoreIntegrationSuite) AfterTest(suiteName, testName string) {
	truncate(cst)
}

func (cst *clientStoreIntegrationSuite) TestCreateClientSuccess() {
	cl := createNewClient(cst.T())

	_, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestCreateClientFailureWhenRecordsAreDuplicate() {
	cl := createNewClient(cst.T())

	_, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	cl, err = client.NewClientBuilder().
		Name(name).
		AccessTokenTTL(accessTokenTTL).
		SessionTTL(sessionTTL).
		MaxActiveSessions(maxActiveSessions).
		PrivateKey(priKey).
		Build()

	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(cst.ctx, cl)
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestRevokeClientSuccess() {
	cl := createNewClient(cst.T())

	secret, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	var id string
	err = cst.db.QueryRow("select id from clients where secret=$1", secret).Scan(&id)
	require.NoError(cst.T(), err)

	_, err = cst.store.RevokeClient(cst.ctx, id)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestRevokeClientFailure() {
	_, err := cst.store.RevokeClient(cst.ctx, id)
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientSuccess() {
	cl := createNewClient(cst.T())

	secret, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(cst.ctx, name, secret)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientFailureWhenRecordIsNotPresent() {
	_, err := cst.store.GetClient(cst.ctx, name, secret)
	require.Error(cst.T(), err)
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(clientStoreIntegrationSuite))
}

func truncate(cst *clientStoreIntegrationSuite) {
	_, err := cst.db.Exec("TRUNCATE clients")
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.cache.FlushAll(cst.ctx).Err())
}
