package database

import (
	"context"
	"database/sql"
	"time"
)

type SQLDatabase interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Close() error
}

//TODO: WRAPS THE ERROR BELOW
//TODO: FIX, USE TIMEOUT FROM PARAMS
type pgDatabase struct {
	timeout int
	db      *sql.DB
}

func (pdb *pgDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	return pdb.db.QueryContext(ctx, query, args...)
}

func (pdb *pgDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	return pdb.db.QueryRowContext(ctx, query, args...)
}

func (pdb *pgDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	return pdb.db.ExecContext(ctx, query, args...)
}

func (pdb *pgDatabase) Close() error {
	return pdb.db.Close()
}

func NewSQLDatabase(db *sql.DB, queryTimeout int) SQLDatabase {
	return &pgDatabase{
		db:      db,
		timeout: queryTimeout,
	}
}
