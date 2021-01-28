package database

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"identification-service/pkg/config"
	"identification-service/pkg/liberr"
	"path/filepath"
	"strings"
)

const (
	cutSet       = "file://"
	databaseName = "postgres"
)

type Migrator interface {
	Migrate() error
	Rollback() error
}

type pgMigrator struct {
	cfg config.MigrationConfig
	mg  *migrate.Migrate
}

func (pg *pgMigrator) Migrate() error {
	return wrap(pg.mg.Up(), "Migrator.Migrate")
}

func (pg *pgMigrator) Rollback() error {
	return wrap(pg.mg.Steps(pg.cfg.RollbackSteps()), "Migrator.Rollback")
}

func wrap(err error, name string) error {
	if err == nil || err == migrate.ErrNoChange {
		return nil
	}

	return liberr.WithOp(liberr.Operation(name), err)
}

func NewMigrator(cfg config.MigrationConfig, db *sql.DB) (Migrator, error) {
	newMigrate, err := newMigrate(cfg, db)
	if err != nil {
		return nil, liberr.WithOp("NewMigrator", err)
	}

	return &pgMigrator{
		cfg: cfg,
		mg:  newMigrate,
	}, nil
}

func newMigrate(cfg config.MigrationConfig, db *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	sourcePath, err := getSourcePath(cfg.Path())
	if err != nil {
		return nil, err
	}

	return migrate.NewWithDatabaseInstance(sourcePath, databaseName, driver)
}

func getSourcePath(directory string) (string, error) {
	directory = strings.TrimLeft(directory, cutSet)

	absPath, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", cutSet, absPath), nil
}
