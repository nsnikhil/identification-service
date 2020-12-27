package database

import (
	"context"
	"database/sql"
	"identification-service/pkg/liberr"
	"time"
)

type SQLDatabase interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Close() error
}

//TODO: WRAPS THE ERROR BELOW
//TODO: FIX NOT CALLING CANCEL
type pgDatabase struct {
	timeout time.Duration
	db      *sql.DB
}

func (pdb *pgDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	//ctx, cancel := context.WithTimeout(ctx, pdb.timeout)
	//defer cancel()

	res, err := pdb.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, liberr.WithOp("SQLDatabase.QueryContext", err)
	}

	return res, nil
}

func (pdb *pgDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	//ctx, cancel := context.WithTimeout(ctx, pdb.timeout)
	//defer cancel()

	res := pdb.db.QueryRowContext(ctx, query, args...)
	return res
}

func (pdb *pgDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	//ctx, cancel := context.WithTimeout(ctx, pdb.timeout)
	//defer cancel()

	res, err := pdb.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, liberr.WithOp("SQLDatabase.ExecContext", err)
	}

	return res, nil
}

func (pdb *pgDatabase) Close() error {
	return pdb.db.Close()
}

func NewSQLDatabase(db *sql.DB, queryTimeout int) SQLDatabase {
	return &pgDatabase{
		db:      db,
		timeout: time.Millisecond * time.Duration(queryTimeout),
	}
}
