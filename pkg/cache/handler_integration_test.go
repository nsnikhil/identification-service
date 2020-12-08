// build integration_test

package cache_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/cache"
	"identification-service/pkg/config"
	"testing"
)

func TestGetCacheSuccess(t *testing.T) {
	cfg := config.NewConfig("../../local.env").CacheConfig()

	_, err := cache.NewHandler(cfg).GetCache()
	assert.Nil(t, err)
}

func TestGetCacheFailure(t *testing.T) {
	_, err := cache.NewHandler(config.CacheConfig{}).GetCache()
	assert.Error(t, err)
}
