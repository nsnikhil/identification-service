package database_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
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

func (sts *sqlDBTestSuite) TestQueryContextSuccess() {
	rows, err := sts.db.QueryContext(sts.ctx, `SELECT version()`)
	require.NoError(sts.T(), err)

	var data interface{}

	for rows.Next() {
		require.NoError(sts.T(), rows.Scan(&data))
	}

	require.NotNil(sts.T(), data)
}

func (sts *sqlDBTestSuite) TestQueryContextFailure() {
	rows, err := sts.db.QueryContext(sts.ctx, fmt.Sprintf("SELECT * FROM %s", test.RandString(8)))
	require.Error(sts.T(), err)
	assert.Nil(sts.T(), rows)
}

func (sts *sqlDBTestSuite) TestQueryRowContextSuccess() {
	row := sts.db.QueryRowContext(sts.ctx, `SELECT version()`)
	require.NoError(sts.T(), row.Err())

	var data interface{}
	require.NoError(sts.T(), row.Scan(&data))

	require.NotNil(sts.T(), data)
}

func (sts *sqlDBTestSuite) TestQueryRowContextFailure() {
	row := sts.db.QueryRowContext(sts.ctx, fmt.Sprintf("SELECT * FROM %s", test.RandString(8)))
	require.Error(sts.T(), row.Err())
}

func (sts *sqlDBTestSuite) TestExecContextSuccess() {
	tableName := test.RandString(8)

	_, err := sts.db.ExecContext(sts.ctx, fmt.Sprintf(`create table if not exists %s (id serial primary key)`, tableName))
	require.NoError(sts.T(), err)

	_, err = sts.db.ExecContext(sts.ctx, fmt.Sprintf(`drop table if exists %s`, tableName))
	require.NoError(sts.T(), err)
}

func (sts *sqlDBTestSuite) TestExecContextFailure() {
	_, err := sts.db.ExecContext(sts.ctx, fmt.Sprintf(`drop table %s`, test.RandString(8)))
	require.Error(sts.T(), err)
}

func TestSQLDatabase(t *testing.T) {
	suite.Run(t, new(sqlDBTestSuite))
}
