package database

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/mock"
)

type MockHandler struct {
	mock.Mock
}

func (mock *MockHandler) GetDB() (*sql.DB, error) {
	args := mock.Called()
	return args.Get(0).(*sql.DB), args.Error(1)
}

type MockSQLDatabase struct {
	mock.Mock
}

//func (mock *MockSQLDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
//	ag := mock.Called(ctx, opts)
//	return ag.Get(0).(*sql.Tx), ag.Error(1)
//}

func (mock *MockSQLDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ag := mock.Called(ctx, query, args)
	return ag.Get(0).(*sql.Rows), ag.Error(1)
}

func (mock *MockSQLDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ag := mock.Called(ctx, query, args)
	return ag.Get(0).(*sql.Row)
}

func (mock *MockSQLDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ag := mock.Called(ctx, query, args)
	return ag.Get(0).(sql.Result), ag.Error(1)
}

func (mock *MockSQLDatabase) Close() error {
	args := mock.Called()
	return args.Error(0)
}
