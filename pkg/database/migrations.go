package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"identification-service/pkg/config"
	"path/filepath"
	"strings"
)

const (
	rollBackStep = -1
	cutSet       = "file://"
	databaseName = "postgres"
)

// TODO: SHOULD IT RETURN LIB ERROR
func RunMigrations(configFile string) {
	newMigrate, err := newMigrate(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := newMigrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return
		}
		fmt.Println(err)
		return
	}
}

// TODO: SHOULD IT RETURN LIB ERROR
func RollBackMigrations(configFile string) {
	newMigrate, err := newMigrate(configFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := newMigrate.Steps(rollBackStep); err != nil {
		if err == migrate.ErrNoChange {
			return
		}
	}
}

func newMigrate(configFile string) (*migrate.Migrate, error) {
	cfg := config.NewConfig(configFile)

	dbHandler := NewHandler(cfg.DatabaseConfig())

	db, err := dbHandler.GetDB()
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	sourcePath, err := getSourcePath(cfg.MigrationPath())
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
