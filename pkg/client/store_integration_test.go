package client_test

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/test"
	"testing"
)

type clientStoreIntegrationSuite struct {
	suite.Suite
	db          database.SQLDatabase
	cache       *redis.Client
	store       client.Store
	ctx         context.Context
	cfg         config.ClientConfig
	defaultData map[string]interface{}
}

func (cst *clientStoreIntegrationSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")
	cst.cfg = cfg.ClientConfig()
	cst.db = test.NewDB(cst.T(), cfg)
	cst.cache = test.NewCache(cst.T(), cfg)
	cst.store = client.NewStore(cst.db, cst.cache)
	cst.ctx = context.Background()
	cst.defaultData = map[string]interface{}{}
}

func (cst *clientStoreIntegrationSuite) TearDownSuite() {
	truncate(cst)
}

func (cst *clientStoreIntegrationSuite) TestCreateClientSuccess() {
	cl, err := test.NewClient(cst.cfg, cst.defaultData)

	_, err = cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestCreateClientFailureWhenRecordsAreDuplicate() {
	cl, err := test.NewClient(cst.cfg, cst.defaultData)

	_, err = cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(cst.ctx, cl)
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestRevokeClientSuccess() {
	cl, err := test.NewClient(cst.cfg, cst.defaultData)

	secret, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	var id string
	err = cst.db.QueryRowContext(cst.ctx, "select id from clients where secret=$1", secret).Scan(&id)
	require.NoError(cst.T(), err)

	_, err = cst.store.RevokeClient(cst.ctx, id)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestRevokeClientFailure() {
	_, err := cst.store.RevokeClient(cst.ctx, test.NewUUID())
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientSuccess() {
	cl, err := test.NewClient(cst.cfg, cst.defaultData)

	secret, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(cst.ctx, cl.Name, secret)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientFromCacheSuccess() {
	cl, err := test.NewClient(cst.cfg, cst.defaultData)

	secret, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(cst.ctx, cl.Name, secret)
	require.NoError(cst.T(), err)

	_, err = cst.db.ExecContext(cst.ctx, "TRUNCATE clients")
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(cst.ctx, cl.Name, secret)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientFromCacheFailureWhenSecretIsInvalid() {
	cl, err := test.NewClient(cst.cfg, cst.defaultData)

	secret, err := cst.store.CreateClient(cst.ctx, cl)
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(cst.ctx, cl.Name, secret)
	require.NoError(cst.T(), err)

	_, err = cst.db.ExecContext(cst.ctx, "TRUNCATE clients")
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(cst.ctx, cl.Name, "invalid")
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientFailureWhenRecordIsNotPresent() {
	_, err := cst.store.GetClient(cst.ctx, test.RandString(8), test.NewUUID())
	require.Error(cst.T(), err)
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(clientStoreIntegrationSuite))
}

func truncate(cst *clientStoreIntegrationSuite) {
	_, err := cst.db.ExecContext(cst.ctx, "TRUNCATE clients")
	require.NoError(cst.T(), err)

	require.NoError(cst.T(), cst.cache.FlushAll(cst.ctx).Err())
}
