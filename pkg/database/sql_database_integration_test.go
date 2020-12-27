package database_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/test"
	"testing"
)

type sqlDBTestSuite struct {
	db  database.SQLDatabase
	ctx context.Context
	suite.Suite
}

func (sts *sqlDBTestSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")
	sts.db = test.NewDB(sts.T(), cfg)
	sts.ctx = context.Background()
}

func (sts *sqlDBTestSuite) AfterTest(suiteName, testName string) {
}

func (sts *sqlDBTestSuite) TearDownSuite() {
	require.NoError(sts.T(), sts.db.Close())
}

func (sts *sqlDBTestSuite) TestQueryContext() {
	rows, err := sts.db.QueryContext(sts.ctx, `SELECT version()`)
	require.NoError(sts.T(), err)

	var data interface{}

	for rows.Next() {
		require.NoError(sts.T(), rows.Scan(&data))
	}

	require.NotNil(sts.T(), data)
}

func (sts *sqlDBTestSuite) TestQueryRowContext() {
	row := sts.db.QueryRowContext(sts.ctx, `SELECT version()`)
	require.NoError(sts.T(), row.Err())

	var data interface{}
	require.NoError(sts.T(), row.Scan(&data))

	require.NotNil(sts.T(), data)
}

func (sts *sqlDBTestSuite) TestExecContext() {
	_, err := sts.db.ExecContext(sts.ctx, `create table if not exists temp (id serial primary key)`)
	require.NoError(sts.T(), err)

	_, err = sts.db.ExecContext(sts.ctx, `drop table if exists temp`)
	require.NoError(sts.T(), err)
}

func TestSQLDatabase(t *testing.T) {
	suite.Run(t, new(sqlDBTestSuite))
}
