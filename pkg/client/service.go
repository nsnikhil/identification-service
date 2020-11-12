package client

import (
	"database/sql"
	"identification-service/pkg/client/internal"
	"identification-service/pkg/liberr"
)

const invalidTTL = -1

type Service interface {
	CreateClient(name string, accessTokenTTL int, sessionTTL int) (string, error)
	RevokeClient(id string) error
	GetClientTTL(name, secret string) (int, int, error)
	ValidateClientCredentials(name, secret string) error
}

type clientService struct {
	store internal.Store
}

func (cs *clientService) CreateClient(name string, accessTokenTTL int, sessionTTL int) (string, error) {
	cl, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	if err != nil {
		return "", liberr.WithOp("Service.CreateClient", err)
	}

	id, err := cs.store.CreateClient(cl)
	if err != nil {
		return "", liberr.WithOp("Service.CreateClient", err)
	}

	return id, nil
}

//TODO: SHOULD IT RETURN THE UPDATE COUNT ?
func (cs *clientService) RevokeClient(id string) error {
	_, err := cs.store.RevokeClient(id)
	if err != nil {
		return liberr.WithOp("Service.RevokeClient", err)
	}

	return nil
}

func (cs *clientService) GetClientTTL(name, secret string) (int, int, error) {
	client, err := cs.store.GetClient(name, secret)
	if err != nil {
		return invalidTTL, invalidTTL, liberr.WithOp("Service.GetClientTTL", err)
	}

	return client.AccessTokenTTL(), client.SessionTTL(), nil
}

func (cs *clientService) ValidateClientCredentials(name, secret string) error {
	_, err := cs.store.GetClient(name, secret)
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

func NewService(db *sql.DB) Service {
	return &clientService{
		store: internal.NewStore(db),
	}
}
