package test

import (
	"database/sql"
	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/cache"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"testing"
)

var sqlDB *sql.DB
var db database.SQLDatabase
var redisClient *redis.Client
var cs sarama.Consumer

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

func NewConsumer(t *testing.T, cfg config.Config) sarama.Consumer {
	if cs != nil {
		return cs
	}

	kcs, err := sarama.NewConsumer(cfg.KafkaConfig().Addresses(), sarama.NewConfig())
	require.NoError(t, err)

	cs = kcs

	return cs
}
