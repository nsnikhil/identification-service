package client

import (
	"context"
	"encoding/base64"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/liberr"
	"time"
)

const invalidTTL = -1

type Service interface {
	CreateClient(ctx context.Context, name string, accessTokenTTL, sessionTTL, maxActiveSessions int) (string, string, error)
	RevokeClient(ctx context.Context, id string) error
	GetClientTTL(ctx context.Context, name, secret string) (int, int, error)
	ValidateClientCredentials(ctx context.Context, name, secret string) error
}

type clientService struct {
	keyGenerator libcrypto.Ed25519Generator
	store        Store
}

func (cs *clientService) CreateClient(ctx context.Context, name string, accessTokenTTL, sessionTTL, maxActiveSessions int) (string, string, error) {
	pubKey, priKey, err := cs.keyGenerator.Generate()
	if err != nil {
		return "", "", liberr.WithOp("Service.CreateClient", err)
	}

	cl, err := NewClientBuilder().
		Name(name).
		AccessTokenTTL(accessTokenTTL).
		SessionTTL(sessionTTL).
		MaxActiveSessions(maxActiveSessions).
		PrivateKey(priKey).
		Build()

	if err != nil {
		return "", "", liberr.WithOp("Service.CreateClient", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	id, err := cs.store.CreateClient(ctx, cl)
	if err != nil {
		return "", "", liberr.WithOp("Service.CreateClient", err)
	}

	//TODO: PULL ENCODING IN SEPARATE PACKAGE
	return base64.RawStdEncoding.EncodeToString(pubKey), id, nil
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

	return client.accessTokenTTL, client.sessionTTL, nil
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

func NewService(store Store, keyGenerator libcrypto.Ed25519Generator) Service {
	return &clientService{
		keyGenerator: keyGenerator,
		store:        store,
	}
}
