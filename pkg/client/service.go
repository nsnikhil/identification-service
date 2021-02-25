package client

import (
	"context"
	"encoding/base64"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
	"identification-service/pkg/libcrypto"
)

type Service interface {
	CreateClient(ctx context.Context, name string, accessTokenTTL, sessionTTL, maxActiveSessions int, sessionStrategy string) (string, string, error)
	RevokeClient(ctx context.Context, id string) error
	GetClient(ctx context.Context, name, secret string) (Client, error)
}

type clientService struct {
	keyGenerator libcrypto.Ed25519Generator
	cfg          config.ClientConfig
	store        Store
}

func (cs *clientService) CreateClient(
	ctx context.Context,
	name string,
	accessTokenTTL,
	sessionTTL,
	maxActiveSessions int,
	sessionStrategy string,
) (string, string, error) {

	pubKey, priKey, err := cs.keyGenerator.Generate()
	if err != nil {
		return "", "", erx.WithArgs(erx.Operation("Service.CreateClient"), err)
	}

	cl, err := NewClientBuilder(cs.cfg).
		Name(name).
		AccessTokenTTL(accessTokenTTL).
		SessionTTL(sessionTTL).
		MaxActiveSessions(maxActiveSessions).
		SessionStrategy(sessionStrategy).
		PrivateKey(priKey).
		Build()

	if err != nil {
		return "", "", erx.WithArgs(erx.Operation("Service.CreateClient"), err)
	}

	id, err := cs.store.CreateClient(ctx, cl)
	if err != nil {
		return "", "", erx.WithArgs(erx.Operation("Service.CreateClient"), err)
	}

	//TODO: PULL ENCODING IN SEPARATE PACKAGE
	return base64.RawStdEncoding.EncodeToString(pubKey), id, nil
}

//TODO: SHOULD IT RETURN THE UPDATE COUNT ?
func (cs *clientService) RevokeClient(ctx context.Context, id string) error {
	_, err := cs.store.RevokeClient(ctx, id)
	if err != nil {
		return erx.WithArgs(erx.Operation("Service.RevokeClient"), err)
	}

	return nil
}

func (cs *clientService) GetClient(ctx context.Context, name, secret string) (Client, error) {
	client, err := cs.store.GetClient(ctx, name, secret)
	if err != nil {
		return Client{}, erx.WithArgs(erx.Operation("Service.GetClient"), err)
	}

	return client, nil
}

func NewService(cfg config.ClientConfig, store Store, keyGenerator libcrypto.Ed25519Generator) Service {
	return &clientService{
		keyGenerator: keyGenerator,
		store:        store,
		cfg:          cfg,
	}
}
