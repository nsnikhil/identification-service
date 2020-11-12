package internal

import (
	"errors"
	"fmt"
	"identification-service/pkg/liberr"
	"identification-service/pkg/util"
	"time"
)

type Client struct {
	id             string
	name           string
	secret         string
	revoked        bool
	accessTokenTTL int
	sessionTTL     int
	createdAt      time.Time
	updatedAt      time.Time
}

func (c Client) AccessTokenTTL() int {
	return c.accessTokenTTL
}

func (c Client) SessionTTL() int {
	return c.sessionTTL
}

type ClientBuilder struct {
	id             string
	name           string
	secret         string
	revoked        bool
	accessTokenTTL int
	sessionTTL     int
	createdAt      time.Time
	updatedAt      time.Time

	err error
}

func (cb *ClientBuilder) ID(id string) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	if !util.IsValidUUID(id) {
		cb.err = fmt.Errorf("invalid user id %s", id)
		return cb
	}

	cb.id = id
	return cb
}

func (cb *ClientBuilder) Name(name string) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	if len(name) == 0 {
		cb.err = errors.New("name cannot be empty")
		return cb
	}

	cb.name = name
	return cb
}

func (cb *ClientBuilder) Secret(secret string) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	if !util.IsValidUUID(secret) {
		cb.err = fmt.Errorf("invalid client secret %s", secret)
		return cb
	}

	cb.id = secret
	return cb
}

func (cb *ClientBuilder) AccessTokenTTL(accessTokenTTL int) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	if accessTokenTTL < 1 {
		cb.err = errors.New("access token ttl cannot be less than one")
		return cb
	}

	cb.accessTokenTTL = accessTokenTTL
	return cb
}

func (cb *ClientBuilder) SessionTTL(sessionTTL int) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	if sessionTTL < 1 {
		cb.err = errors.New("session ttl cannot be less than one")
		return cb
	}

	cb.sessionTTL = sessionTTL
	return cb
}

func (cb *ClientBuilder) CreatedAt(createdAt time.Time) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	cb.createdAt = createdAt
	return cb
}

func (cb *ClientBuilder) UpdatedAt(updatedAt time.Time) *ClientBuilder {
	if cb.err != nil {
		return cb
	}

	cb.updatedAt = updatedAt
	return cb
}

func (cb *ClientBuilder) Build() (Client, error) {
	if cb.err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("ClientBuilder.Build"), liberr.ValidationError, cb.err)
	}

	if err := validateArgs(cb.name, cb.accessTokenTTL, cb.sessionTTL); err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("ClientBuilder.Build"), liberr.ValidationError, err)
	}

	return Client{
		id:             cb.id,
		name:           cb.name,
		secret:         cb.secret,
		accessTokenTTL: cb.accessTokenTTL,
		sessionTTL:     cb.sessionTTL,
		createdAt:      cb.createdAt,
		updatedAt:      cb.updatedAt,
	}, nil
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{}
}

func validateArgs(name string, accessTokenTTL int, sessionTTL int) error {
	if len(name) == 0 {
		return errors.New("client name cannot be empty")
	}

	if accessTokenTTL < 1 {
		return errors.New("access token ttl cannot be less than one")
	}

	if sessionTTL < 1 {
		return errors.New("session ttl cannot be less than one")
	}

	if accessTokenTTL > sessionTTL {
		return errors.New("session ttl cannot be less than access token ttl")
	}

	return nil
}
