package client

import (
	"errors"
	"fmt"
	"identification-service/pkg/liberr"
	"identification-service/pkg/util"
	"time"
)

type Client struct {
	id                string
	name              string
	secret            string
	revoked           bool
	accessTokenTTL    int
	sessionTTL        int
	maxActiveSessions int
	privateKey        []byte
	createdAt         time.Time
	updatedAt         time.Time
}

type Builder struct {
	id                string
	name              string
	secret            string
	revoked           bool
	accessTokenTTL    int
	sessionTTL        int
	maxActiveSessions int
	privateKey        []byte
	createdAt         time.Time
	updatedAt         time.Time

	err error
}

func (b *Builder) ID(id string) *Builder {
	if b.err != nil {
		return b
	}

	if !util.IsValidUUID(id) {
		b.err = fmt.Errorf("invalid client id %s", id)
		return b
	}

	b.id = id
	return b
}

func (b *Builder) Name(name string) *Builder {
	if b.err != nil {
		return b
	}

	if len(name) == 0 {
		b.err = errors.New("name cannot be empty")
		return b
	}

	b.name = name
	return b
}

func (b *Builder) Secret(secret string) *Builder {
	if b.err != nil {
		return b
	}

	if !util.IsValidUUID(secret) {
		b.err = fmt.Errorf("invalid client secret %s", secret)
		return b
	}

	b.id = secret
	return b
}

func (b *Builder) AccessTokenTTL(accessTokenTTL int) *Builder {
	if b.err != nil {
		return b
	}

	if accessTokenTTL < 1 {
		b.err = errors.New("access token ttl cannot be less than one")
		return b
	}

	b.accessTokenTTL = accessTokenTTL
	return b
}

func (b *Builder) SessionTTL(sessionTTL int) *Builder {
	if b.err != nil {
		return b
	}

	if sessionTTL < 1 {
		b.err = errors.New("session ttl cannot be less than one")
		return b
	}

	b.sessionTTL = sessionTTL
	return b
}
func (b *Builder) MaxActiveSessions(maxActiveSessions int) *Builder {
	if b.err != nil {
		return b
	}

	if maxActiveSessions < 1 {
		b.err = errors.New("max active sessions cannot be less than one")
		return b
	}

	b.maxActiveSessions = maxActiveSessions
	return b
}

func (b *Builder) PrivateKey(privateKey []byte) *Builder {
	if b.err != nil {
		return b
	}

	if len(privateKey) == 0 {
		b.err = errors.New("private key cannot be empty")
		return b
	}

	b.privateKey = privateKey
	return b
}

func (b *Builder) CreatedAt(createdAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	b.createdAt = createdAt
	return b
}

func (b *Builder) UpdatedAt(updatedAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	b.updatedAt = updatedAt
	return b
}

func (b *Builder) Build() (Client, error) {
	if b.err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("ClientBuilder.Build"), liberr.ValidationError, b.err)
	}

	if err := validateArgs(b); err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("ClientBuilder.Build"), liberr.ValidationError, err)
	}

	return Client{
		id:                b.id,
		name:              b.name,
		secret:            b.secret,
		accessTokenTTL:    b.accessTokenTTL,
		sessionTTL:        b.sessionTTL,
		maxActiveSessions: b.maxActiveSessions,
		privateKey:        b.privateKey,
		createdAt:         b.createdAt,
		updatedAt:         b.updatedAt,
	}, nil
}

func NewClientBuilder() *Builder {
	return &Builder{}
}

//TODO: THIS IS CURRENTLY REPEATED BECAUSE USING BUILDER SOMEONE MIGHT NOT SET THESE VALUES
func validateArgs(b *Builder) error {
	if len(b.name) == 0 {
		return errors.New("client name cannot be empty")
	}

	if b.accessTokenTTL < 1 {
		return errors.New("access token ttl cannot be less than one")
	}

	if b.sessionTTL < 1 {
		return errors.New("session ttl cannot be less than one")
	}

	if b.accessTokenTTL > b.sessionTTL {
		return errors.New("session ttl cannot be less than access token ttl")
	}

	if b.maxActiveSessions < 1 {
		return errors.New("max active sessions cannot be less than one")
	}

	if b.privateKey == nil || len(b.privateKey) == 0 {
		return errors.New("private key cannot be empty")
	}

	return nil
}
