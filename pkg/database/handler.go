package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
	"time"
)

type Handler interface {
	GetDB() (*sql.DB, error)
}

type sqlDBHandler struct {
	cfg config.DatabaseConfig
}

func (dbh *sqlDBHandler) GetDB() (*sql.DB, error) {
	db, err := sql.Open(dbh.cfg.DriverName(), dbh.cfg.Source())
	if err != nil {
		return nil, erx.WithArgs(erx.Operation("Handler.GetDB"), erx.SeverityError, err)
	}

	db.SetMaxOpenConns(dbh.cfg.MaxOpenConnections())
	db.SetMaxIdleConns(dbh.cfg.IdleConnections())
	db.SetConnMaxLifetime(time.Minute * time.Duration(dbh.cfg.ConnectionMaxLifetime()))

	if err := db.Ping(); err != nil {
		return nil, erx.WithArgs(erx.Operation("Handler.GetDB"), erx.SeverityError, err)
	}

	return db, nil
}

func NewHandler(cfg config.DatabaseConfig) Handler {
	return &sqlDBHandler{
		cfg: cfg,
	}
}
