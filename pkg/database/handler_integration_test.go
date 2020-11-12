package database_test

import (
	"github.com/stretchr/testify/require"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"testing"
)

func TestGetDBSuccess(t *testing.T) {
	cfg := config.NewConfig("../../local.env").DatabaseConfig()

	handler := database.NewHandler(cfg)

	_, err := handler.GetDB()
	require.NoError(t, err)
}

func TestGetDBFailure(t *testing.T) {
	cfg := config.DatabaseConfig{}

	handler := database.NewHandler(cfg)

	_, err := handler.GetDB()
	require.Error(t, err)
}
