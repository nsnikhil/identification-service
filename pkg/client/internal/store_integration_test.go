// build integration_test

package internal_test

import (
	"database/sql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client/internal"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"testing"
)

type clientStoreIntegrationSuite struct {
	suite.Suite
	db    *sql.DB
	store internal.Store
}

func (cst *clientStoreIntegrationSuite) SetupSuite() {
	cst.db = getDB(cst.T())
	cst.store = internal.NewStore(cst.db)
}

func (cst *clientStoreIntegrationSuite) AfterTest(suiteName, testName string) {
	truncate(cst.T(), cst.db)
}

func (cst *clientStoreIntegrationSuite) TestCreateClientSuccess() {
	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(cl)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestCreateClientFailureWhenRecordsAreDuplicate() {
	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(cl)
	require.NoError(cst.T(), err)

	cl, err = internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	_, err = cst.store.CreateClient(cl)
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestRevokeClientSuccess() {
	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	secret, err := cst.store.CreateClient(cl)
	require.NoError(cst.T(), err)

	var id string
	err = cst.db.QueryRow("select id from clients where secret=$1", secret).Scan(&id)
	require.NoError(cst.T(), err)

	_, err = cst.store.RevokeClient(id)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestRevokeClientFailure() {
	_, err := cst.store.RevokeClient(id)
	require.Error(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientSuccess() {
	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(cst.T(), err)

	secret, err := cst.store.CreateClient(cl)
	require.NoError(cst.T(), err)

	_, err = cst.store.GetClient(name, secret)
	require.NoError(cst.T(), err)
}

func (cst *clientStoreIntegrationSuite) TestGetClientFailureWhenRecordIsNotPresent() {
	_, err := cst.store.GetClient(name, secret)
	require.Error(cst.T(), err)
}

func TestStoreIntegration(t *testing.T) {
	suite.Run(t, new(clientStoreIntegrationSuite))
}

func truncate(t *testing.T, db *sql.DB) {
	_, err := db.Exec("TRUNCATE clients")
	require.NoError(t, err)
}

func getDB(t *testing.T) *sql.DB {
	cfg := config.NewConfig("../../../local.env")

	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	require.NoError(t, err)

	return db
}
