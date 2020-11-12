package config

import "fmt"

type CacheConfig struct {
	host     string
	port     int
	username string
	password string
	database int
}

func newCacheConfig() CacheConfig {
	return CacheConfig{
		host:     getString("CACHE_HOST"),
		port:     getInt("CACHE_PORT"),
		username: getString("CACHE_USER_NAME"),
		password: getString("CACHE_PASSWORD"),
		database: getInt("CACHE_DATABASE"),
	}
}

func (cc CacheConfig) Address() string {
	return fmt.Sprintf("%s:%d", cc.host, cc.port)
}

func (cc CacheConfig) UserName() string {
	return cc.username
}

func (cc CacheConfig) Password() string {
	return cc.password
}

func (cc CacheConfig) Database() int {
	return cc.database
}
