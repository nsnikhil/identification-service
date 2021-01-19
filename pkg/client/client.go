package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"identification-service/pkg/config"
	"identification-service/pkg/liberr"
	"identification-service/pkg/util"
	"time"
)

type ctxKey string

var clientCtxKey ctxKey = "clientCtxKey"

type Client struct {
	internalClient
}

type internalClient struct {
	Id                  string
	Name                string
	Secret              string
	Revoked             bool
	AccessTokenTTL      int
	SessionTTL          int
	MaxActiveSessions   int
	SessionStrategyName string
	PrivateKey          []byte
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (cl Client) IsRevoked() bool {
	return cl.Revoked
}

func (cl Client) AccessTokenTTL() int {
	return cl.internalClient.AccessTokenTTL
}

func (cl Client) SessionStrategyName() string {
	return cl.internalClient.SessionStrategyName
}

func (cl Client) SessionTTL() int {
	return cl.internalClient.SessionTTL
}

func (cl Client) MaxActiveSessions() int {
	return cl.internalClient.MaxActiveSessions
}

type Builder struct {
	id                  string
	name                string
	secret              string
	revoked             bool
	accessTokenTTL      int
	sessionTTL          int
	maxActiveSessions   int
	sessionStrategyName string
	privateKey          []byte
	createdAt           time.Time
	updatedAt           time.Time

	err error
	cfg config.ClientConfig
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

func (b *Builder) Revoked(revoked bool) *Builder {
	if b.err != nil {
		return b
	}

	b.revoked = revoked
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

func (b *Builder) SessionStrategy(sessionStrategyName string) *Builder {
	if b.err != nil {
		return b
	}

	if len(sessionStrategyName) == 0 {
		b.err = errors.New("session strategy name cannot be empty")
		return b
	}

	if !b.cfg.Strategies()[sessionStrategyName] {
		b.err = fmt.Errorf("invalid session strategy %s", sessionStrategyName)
		return b
	}

	b.sessionStrategyName = sessionStrategyName
	return b
}

func (b *Builder) PrivateKey(privateKey []byte) *Builder {
	if b.err != nil {
		return b
	}

	if privateKey == nil || len(privateKey) == 0 {
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

	if createdAt == (time.Time{}) {
		b.err = errors.New("invalid created at time")
		return b
	}

	b.createdAt = createdAt
	return b
}

func (b *Builder) UpdatedAt(updatedAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	if updatedAt == (time.Time{}) {
		b.err = errors.New("invalid updated at time")
		return b
	}

	b.updatedAt = updatedAt
	return b
}

func (b *Builder) Build() (Client, error) {
	if b.err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("ClientBuilder.Build"), liberr.ValidationError, b.err)
	}

	if err := validateArgs(b.name, b.accessTokenTTL, b.sessionTTL, b.maxActiveSessions, b.sessionStrategyName, b.privateKey); err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("ClientBuilder.Build"), liberr.ValidationError, err)
	}

	return Client{
		internalClient{
			Id:                  b.id,
			Name:                b.name,
			Secret:              b.secret,
			Revoked:             b.revoked,
			AccessTokenTTL:      b.accessTokenTTL,
			SessionTTL:          b.sessionTTL,
			MaxActiveSessions:   b.maxActiveSessions,
			SessionStrategyName: b.sessionStrategyName,
			PrivateKey:          b.privateKey,
			CreatedAt:           b.createdAt,
			UpdatedAt:           b.updatedAt,
		},
	}, nil
}

func NewClientBuilder(cfg config.ClientConfig) *Builder {
	return &Builder{
		cfg: cfg,
	}
}

func WithContext(ctx context.Context, cl Client) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("base context is nil")
	}

	err := validateArgs(
		cl.Name,
		cl.internalClient.AccessTokenTTL,
		cl.internalClient.SessionTTL,
		cl.internalClient.MaxActiveSessions,
		cl.internalClient.SessionStrategyName,
		cl.PrivateKey,
	)

	if err != nil {
		return nil, liberr.WithArgs(liberr.Operation("Client.WithContext"), liberr.ValidationError, err)
	}

	return context.WithValue(ctx, clientCtxKey, cl), nil
}

func FromContext(ctx context.Context) (Client, error) {
	res := ctx.Value(clientCtxKey)
	if res == nil {
		return Client{}, liberr.WithOp(
			"Client.FromContext",
			errors.New("client info not present in context"),
		)
	}

	cl, ok := res.(Client)
	if !ok {
		return Client{}, liberr.WithOp(
			"Client.FromContext",
			errors.New("invalid client info"),
		)
	}

	err := validateArgs(
		cl.Name,
		cl.internalClient.AccessTokenTTL,
		cl.internalClient.SessionTTL,
		cl.internalClient.MaxActiveSessions,
		cl.internalClient.SessionStrategyName,
		cl.internalClient.PrivateKey,
	)

	if err != nil {
		return Client{}, liberr.WithArgs(liberr.Operation("Client.WithContext"), liberr.ValidationError, err)
	}

	return cl, nil
}

//TODO: THIS IS CURRENTLY REPEATED BECAUSE USING BUILDER SOMEONE MIGHT NOT SET THESE VALUES
func validateArgs(name string, accessTokenTTL, sessionTTL, maxActiveSessions int, sessionStrategyName string, privateKey []byte) error {
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

	if maxActiveSessions < 1 {
		return errors.New("max active sessions cannot be less than one")
	}

	//TODO: NO VALIDATION ON THE NAME HERE
	if len(sessionStrategyName) == 0 {
		return errors.New("session strategy name cannot be empty")
	}

	if privateKey == nil || len(privateKey) == 0 {
		return errors.New("private key cannot be empty")
	}

	return nil
}

func encode(cl Client) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(cl.internalClient)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func decode(str string) (Client, error) {
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return Client{}, err
	}

	b := bytes.Buffer{}
	b.Write(by)

	d := gob.NewDecoder(&b)

	var cl Client
	err = d.Decode(&cl.internalClient)
	if err != nil {
		return Client{}, err
	}

	return cl, nil
}
