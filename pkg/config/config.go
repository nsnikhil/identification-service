package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config interface {
	HTTPServerConfig() HTTPServerConfig
	DatabaseConfig() DatabaseConfig
	LogConfig() LogConfig
	LogFileConfig() LogFileConfig
	Env() string
	MigrationConfig() MigrationConfig
	PasswordConfig() PasswordConfig
	TokenConfig() TokenConfig
	AuthConfig() AuthConfig
	CacheConfig() CacheConfig
	ClientConfig() ClientConfig
	QueueConfig() QueueConfig
}

type appConfig struct {
	env              string
	migrationConfig  MigrationConfig
	httpServerConfig HTTPServerConfig
	databaseConfig   DatabaseConfig
	logConfig        LogConfig
	logFileConfig    LogFileConfig
	passwordConfig   PasswordConfig
	tokenConfig      TokenConfig
	cacheConfig      CacheConfig
	authConfig       AuthConfig
	clientConfig     ClientConfig
	ampqConfig       QueueConfig
}

func (c appConfig) HTTPServerConfig() HTTPServerConfig {
	return c.httpServerConfig
}

func (c appConfig) DatabaseConfig() DatabaseConfig {
	return c.databaseConfig
}

func (c appConfig) LogConfig() LogConfig {
	return c.logConfig
}

func (c appConfig) LogFileConfig() LogFileConfig {
	return c.logFileConfig
}

func (c appConfig) Env() string {
	return c.env
}

func (c appConfig) MigrationConfig() MigrationConfig {
	return c.migrationConfig
}

func (c appConfig) PasswordConfig() PasswordConfig {
	return c.passwordConfig
}

func (c appConfig) TokenConfig() TokenConfig {
	return c.tokenConfig
}

func (c appConfig) AuthConfig() AuthConfig {
	return c.authConfig
}

func (c appConfig) CacheConfig() CacheConfig {
	return c.cacheConfig
}

func (c appConfig) ClientConfig() ClientConfig {
	return c.clientConfig
}

func (c appConfig) QueueConfig() QueueConfig {
	return c.ampqConfig
}

//TODO: FIGURE OUT OF WAY TO KEEP ONE CONFIG FILE FOR LOCAL AND DOCKER
func NewConfig(configFile string) Config {
	viper.AutomaticEnv()
	viper.SetConfigFile(configFile)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	return appConfig{
		env:              getString("ENV"),
		migrationConfig:  NewMigrationConfig(),
		httpServerConfig: newHTTPServerConfig(),
		databaseConfig:   newDatabaseConfig(),
		logConfig:        newLogConfig(),
		logFileConfig:    newLogFileConfig(),
		passwordConfig:   newPasswordConfig(),
		tokenConfig:      newTokenConfig(),
		authConfig:       newAuthConfig(),
		cacheConfig:      newCacheConfig(),
		clientConfig:     newClientConfig(),
		ampqConfig:       newQueueConfig(),
	}
}
