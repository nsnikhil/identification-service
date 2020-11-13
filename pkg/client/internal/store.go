package internal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis/v8"
	"identification-service/pkg/liberr"
)

const (
	createClient = `insert into clients (name, accesstokenttl, sessionttl) values ($1, $2, $3) returning secret`
	revokeClient = `update clients set revoked=true where id=$1`
	getClient    = `select id, revoked, accesstokenttl, sessionttl from clients where name=$1 and secret=$2`

	secretKey         = `secret`
	revokedKey        = "revoked"
	accessTokenTTLKey = `accessTokenTTLKey`
	sessionTTLKey     = `sessionTTLKey`
)

type Store interface {
	CreateClient(ctx context.Context, client Client) (string, error)
	RevokeClient(ctx context.Context, id string) (int64, error)
	GetClient(ctx context.Context, name, secret string) (Client, error)
}

type clientStore struct {
	db    *sql.DB
	cache *redis.Client
}

func (cs *clientStore) CreateClient(ctx context.Context, client Client) (string, error) {
	var secret string

	//TODO: RETURN DIFFERENT ERROR KIND FOR DUPLICATE RECORD
	err := cs.db.QueryRowContext(ctx, createClient, client.name, client.accessTokenTTL, client.sessionTTL).Scan(&secret)
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
		return 0, liberr.WithArgs(liberr.Operation("Store.RevokeClient"), liberr.ResourceNotFound, fmt.Errorf("no client found with id %s", id))
	}

	return c, nil
}

func (cs *clientStore) GetClient(ctx context.Context, name, secret string) (Client, error) {
	var client Client

	row := cs.db.QueryRowContext(ctx, getClient, name, secret)
	if row.Err() != nil {
		return client, liberr.WithOp("Store.GetClient", row.Err())
	}

	err := row.Scan(&client.id, &client.revoked, &client.accessTokenTTL, &client.sessionTTL)
	if err != nil {
		return client, liberr.WithOp("Store.GetClient", err)
	}

	return client, nil
}

//TODO: PICK TTL FROM CONFIG
//TODO: MOVE THIS LOGIC TO CACHE PACKAGE
func saveClientToCache(ctx context.Context, redisClient *redis.Client, name, secret string, revoked bool, accessTokenTTL, sessionTTL int) error {
	res := redisClient.HSet(ctx, name, secretKey, secret, revokedKey, revoked, accessTokenTTLKey, accessTokenTTL, sessionTTLKey, sessionTTL)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

//TODO: MOVE THIS LOGIC TO CACHE PACKAGE
func getClientFromCache(ctx context.Context, redisClient *redis.Client, key string) (map[string]string, error) {
	res := redisClient.HGetAll(ctx, key)
	if res.Err() != nil {
		return nil, res.Err()
	}

	return res.Result()
}

func NewStore(db *sql.DB, cache *redis.Client) Store {
	return &clientStore{
		db:    db,
		cache: cache,
	}
}
