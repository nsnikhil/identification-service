package client

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"identification-service/pkg/database"
	"identification-service/pkg/liberr"
	"time"
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

	row := cs.db.QueryRowContext(
		ctx,
		createClient,
		client.Name,
		client.internalClient.AccessTokenTTL,
		client.internalClient.SessionTTL,
		client.internalClient.MaxActiveSessions,
		client.internalClient.SessionStrategyName,
		client.PrivateKey,
	)

	//TODO: REMOVE THIS HARD CODING
	if row.Err() != nil {
		if pgErr, ok := row.Err().(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return "", liberr.WithArgs(liberr.Operation("Store.CreateClient"), liberr.DuplicateRecordError, row.Err())
			}
		}

		return "", liberr.WithOp("Store.CreateClient", row.Err())
	}

	var secret string

	err := row.Scan(&secret)
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
	if cl, err := fetchFromCache(ctx, cs.cache, name); err == nil {
		return cl, nil
	}

	row := cs.db.QueryRowContext(ctx, getClient, name, secret)
	if row.Err() != nil {
		return Client{}, liberr.WithOp("Store.GetClient", row.Err())
	}

	var client Client
	err := row.Scan(
		&client.Id,
		&client.Revoked,
		&client.internalClient.AccessTokenTTL,
		&client.internalClient.SessionTTL,
		&client.internalClient.MaxActiveSessions,
		&client.internalClient.SessionStrategyName,
		&client.PrivateKey,
	)

	if err != nil {
		return client, liberr.WithOp("Store.GetClient", err)
	}

	//TODO: REFACTOR THIS
	client.Name = name
	client.Secret = secret

	//TODO: HANDLE ERROR
	go updateCache(ctx, cs.cache, client)

	return client, nil
}

func updateCache(ctx context.Context, cache *redis.Client, cl Client) error {
	s, err := encode(cl)
	if err != nil {
		return err
	}

	//TODO: PICK CONFIG FROM TIME
	_, err = cache.Set(ctx, cl.Name, s, time.Hour).Result()
	if err != nil {
		return err
	}

	return nil
}

func fetchFromCache(ctx context.Context, cache *redis.Client, name string) (Client, error) {
	s, err := cache.Get(ctx, name).Result()
	if err != nil {
		return Client{}, err
	}

	cl, err := decode(s)
	if err != nil {
		return Client{}, err
	}

	return cl, err
}

func NewStore(db database.SQLDatabase, cache *redis.Client) Store {
	return &clientStore{
		db:    db,
		cache: cache,
	}
}
