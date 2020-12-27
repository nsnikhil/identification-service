package test

import (
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/cache"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"testing"
)

var db database.SQLDatabase
var redisClient *redis.Client
var channel *amqp.Channel

func NewDB(t *testing.T, cfg config.Config) database.SQLDatabase {
	if db != nil {
		return db
	}

	dbCfg := cfg.DatabaseConfig()

	sqlDB, err := database.NewHandler(dbCfg).GetDB()
	require.NoError(t, err)

	require.NoError(t, sqlDB.Ping())

	db = database.NewSQLDatabase(sqlDB, dbCfg.QueryTTL())

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

func NewChannel(t *testing.T, cfg config.Config) *amqp.Channel {
	if channel != nil {
		return channel
	}

	conn, err := amqp.Dial(cfg.AMPQConfig().Address())
	require.NoError(t, err)

	channel, err = conn.Channel()
	require.NoError(t, err)

	return channel
}
