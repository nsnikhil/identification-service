package client

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"identification-service/pkg/database"
	"identification-service/pkg/liberr"
)

const (
	createClient = `insert into clients (name, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key) values ($1, $2, $3, $4, $5, $6) returning secret`
	revokeClient = `update clients set revoked=true where id=$1`
	getClient    = `select id, revoked, access_token_ttl, session_ttl, max_active_sessions, session_strategy, private_key from clients where name=$1 and secret=$2`
)

type Store interface {
	CreateClient(ctx context.Context, client Client) (string, error)
	RevokeClient(ctx context.Context, id string) (int64, error)
	GetClient(ctx context.Context, name, secret string) (Client, error)
}

type clientStore struct {
	db    database.SQLDatabase
	cache *redis.Client
}

func (cs *clientStore) CreateClient(ctx context.Context, client Client) (string, error) {
	var secret string

	//TODO: RETURN DIFFERENT ERROR KIND FOR DUPLICATE RECORD
	err := cs.db.QueryRowContext(ctx, createClient,
		client.name,
		client.accessTokenTTL,
		client.sessionTTL,
		client.maxActiveSessions,
		client.sessionStrategyName,
		client.privateKey,
	).Scan(&secret)

	if err != nil {
		return "", liberr.WithOp("Store.CreateClient", err)
	}

	return secret, nil
}

func (cs *clientStore) RevokeClient(ctx context.Context, id string) (int64, error) {
	wrap := func(err error) error { return liberr.WithOp("Store.RevokeClient", err) }

	res, err := cs.db.ExecContext(ctx, revokeClient, id)
	if err != nil {
		return 0, wrap(err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return 0, wrap(err)
	}

	if c == 0 {
		return 0, liberr.WithArgs(
			liberr.Operation("Store.RevokeClient"),
			liberr.ResourceNotFound,
			fmt.Errorf("no client found with id %s", id),
		)
	}

	return c, nil
}

func (cs *clientStore) GetClient(ctx context.Context, name, secret string) (Client, error) {
	row := cs.db.QueryRowContext(ctx, getClient, name, secret)
	if row.Err() != nil {
		return Client{}, liberr.WithOp("Store.GetClient", row.Err())
	}

	var client Client
	err := row.Scan(
		&client.id,
		&client.revoked,
		&client.accessTokenTTL,
		&client.sessionTTL,
		&client.maxActiveSessions,
		&client.sessionStrategyName,
		&client.privateKey,
	)

	if err != nil {
		return client, liberr.WithOp("Store.GetClient", err)
	}

	//TODO: REFACTOR THIS
	client.name = name
	client.secret = secret

	return client, nil
}

func NewStore(db database.SQLDatabase, cache *redis.Client) Store {
	return &clientStore{
		db:    db,
		cache: cache,
	}
}
