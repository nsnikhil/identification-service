package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
)

type MockHandler struct {
	mock.Mock
}

func (mock *MockHandler) GetCache() (*redis.Client, error) {
	args := mock.Called()
	return args.Get(0).(*redis.Client), args.Error(1)
}
