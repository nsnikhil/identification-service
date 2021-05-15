package test

import (
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/cache"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/queue"
	"testing"
)

var sqlDB *sql.DB
var db database.SQLDatabase
var redisClient *redis.Client
var qu queue.Queue

func NewSqlDB(t *testing.T, cfg config.Config) *sql.DB {
	if sqlDB != nil {
		return sqlDB
	}

	dbCfg := cfg.DatabaseConfig()

	db, err := database.NewHandler(dbCfg).GetDB()
	require.NoError(t, err)

	sqlDB = db

	return sqlDB
}

func NewDB(t *testing.T, cfg config.Config) database.SQLDatabase {
	if db != nil {
		return db
	}

	sqlDB = NewSqlDB(t, cfg)

	db = database.NewSQLDatabase(sqlDB, cfg.DatabaseConfig().QueryTTL())

	return db
}

func NewCache(t *testing.T, cfg config.Config) *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	var err error

	redisClient, err = cache.NewHandler(cfg.CacheConfig()).GetCache()
	require.NoError(t, err)

	return redisClient
}

func NewQueue(t *testing.T, cfg config.QueueConfig) queue.Queue {
	if qu != nil {
		return qu
	}

	ch, err := queue.NewHandler(cfg).GetChannel()
	require.NoError(t, err)

	qu := queue.NewQueue(ch)

	return qu
}
