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
	MigrationPath() string
	PasswordConfig() PasswordConfig
	TokenConfig() TokenConfig
	AuthConfig() AuthConfig
	CacheConfig() CacheConfig
	PublisherConfig() PublisherConfig
	ConsumerConfig() ConsumerConfig
	AMPQConfig() AMPQConfig
	ClientConfig() ClientConfig
}

type appConfig struct {
	env              string
	migrationPath    string
	httpServerConfig HTTPServerConfig
	databaseConfig   DatabaseConfig
	logConfig        LogConfig
	logFileConfig    LogFileConfig
	passwordConfig   PasswordConfig
	tokenConfig      TokenConfig
	ampqConfig       AMPQConfig
	cacheConfig      CacheConfig
	publisherConfig  PublisherConfig
	consumerConfig   ConsumerConfig
	authConfig       AuthConfig
	clientConfig     ClientConfig
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

func (c appConfig) MigrationPath() string {
	return c.migrationPath
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

func (c appConfig) PublisherConfig() PublisherConfig {
	return c.publisherConfig
}

func (c appConfig) ConsumerConfig() ConsumerConfig {
	return c.consumerConfig
}

func (c appConfig) AMPQConfig() AMPQConfig {
	return c.ampqConfig
}

func (c appConfig) ClientConfig() ClientConfig {
	return c.clientConfig
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
		migrationPath:    getString("MIGRATION_PATH"),
		httpServerConfig: newHTTPServerConfig(),
		databaseConfig:   newDatabaseConfig(),
		logConfig:        newLogConfig(),
		logFileConfig:    newLogFileConfig(),
		passwordConfig:   newPasswordConfig(),
		tokenConfig:      newTokenConfig(),
		authConfig:       newAuthConfig(),
		cacheConfig:      newCacheConfig(),
		publisherConfig:  newPublisherConfig(),
		consumerConfig:   newConsumerConfig(),
		ampqConfig:       newAMPQConfig(),
		clientConfig:     newClientConfig(),
	}
}
