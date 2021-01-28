package config

import "github.com/stretchr/testify/mock"

type MigrationConfig interface {
	Path() string
	RollbackSteps() int
}

type appMigrationConfig struct {
	path          string
	rollbackSteps int
}

func (amc appMigrationConfig) Path() string {
	return amc.path
}

func (amc appMigrationConfig) RollbackSteps() int {
	return amc.rollbackSteps
}

func NewMigrationConfig() MigrationConfig {
	return appMigrationConfig{
		path:          getString("MIGRATION_PATH"),
		rollbackSteps: getInt("MIGRATION_ROLLBACK_STEPS"),
	}
}

type MockMigrationConfig struct {
	mock.Mock
}

func (mock *MockMigrationConfig) Path() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockMigrationConfig) RollbackSteps() int {
	args := mock.Called()
	return args.Int(0)
}
