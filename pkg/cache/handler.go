package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
)

type Handler interface {
	GetCache() (*redis.Client, error)
}

type redisHandler struct {
	cfg config.CacheConfig
}

func (rh *redisHandler) GetCache() (*redis.Client, error) {
	//TODO: MODIFY OTHER CONFIGS
	opt := &redis.Options{
		Addr:     rh.cfg.Address(),
		Username: rh.cfg.UserName(),
		Password: rh.cfg.Password(),
		DB:       rh.cfg.Database(),
	}

	cl := redis.NewClient(opt)

	cmd := cl.Ping(context.Background())
	if cmd.Err() != nil {
		return nil, erx.WithArgs(erx.Operation("Handler.GetCache"), cmd.Err())
	}

	return cl, nil
}

func NewHandler(cfg config.CacheConfig) Handler {
	return &redisHandler{
		cfg: cfg,
	}
}
