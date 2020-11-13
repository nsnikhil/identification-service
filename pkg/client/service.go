package client

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"identification-service/pkg/client/internal"
	"identification-service/pkg/liberr"
	"time"
)

const invalidTTL = -1

type Service interface {
	CreateClient(ctx context.Context, name string, accessTokenTTL int, sessionTTL int) (string, error)
	RevokeClient(ctx context.Context, id string) error
	GetClientTTL(ctx context.Context, name, secret string) (int, int, error)
	ValidateClientCredentials(ctx context.Context, name, secret string) error
}

type clientService struct {
	store internal.Store
}

func (cs *clientService) CreateClient(ctx context.Context, name string, accessTokenTTL int, sessionTTL int) (string, error) {
	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	if err != nil {
		return "", liberr.WithOp("Service.CreateClient", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	id, err := cs.store.CreateClient(ctx, cl)
	if err != nil {
		return "", liberr.WithOp("Service.CreateClient", err)
	}

	return id, nil
}

//TODO: SHOULD IT RETURN THE UPDATE COUNT ?
func (cs *clientService) RevokeClient(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := cs.store.RevokeClient(ctx, id)
	if err != nil {
		return liberr.WithOp("Service.RevokeClient", err)
	}

	return nil
}

func (cs *clientService) GetClientTTL(ctx context.Context, name, secret string) (int, int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	client, err := cs.store.GetClient(ctx, name, secret)
	if err != nil {
		return invalidTTL, invalidTTL, liberr.WithOp("Service.GetClientTTL", err)
	}

	return client.AccessTokenTTL(), client.SessionTTL(), nil
}

func (cs *clientService) ValidateClientCredentials(ctx context.Context, name, secret string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := cs.store.GetClient(ctx, name, secret)
	if err != nil {
		return liberr.WithOp("Service.ValidateClientCredentials", err)
	}

	return nil
}

//TODO: FIGURE OUT A WAY TO REMOVE THIS AS IT IS ONLY NEEDED FOR TEST
func NewInternalService(store internal.Store) Service {
	return &clientService{
		store: store,
	}
}

func NewService(db *sql.DB, cache *redis.Client) Service {
	return &clientService{
		store: internal.NewStore(db, cache),
	}
}
