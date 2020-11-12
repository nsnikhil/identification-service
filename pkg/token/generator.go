package token

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"golang.org/x/crypto/ssh"
	"identification-service/pkg/config"
	"identification-service/pkg/liberr"
	"time"
)

type Generator interface {
	GenerateAccessToken(ttl int, subject string, claims map[string]string) (string, error)
	GenerateRefreshToken() (string, error)
}

type pasetoTokenGenerator struct {
	audience   string
	issuer     string
	privateKey ed25519.PrivateKey
}

func (tg *pasetoTokenGenerator) GenerateAccessToken(ttl int, subject string, claims map[string]string) (string, error) {
	now := time.Now()

	jsonToken := getJSONToken(now, ttl, tg.audience, tg.issuer, subject, claims)

	accessToken, err := paseto.NewV2().Sign(tg.privateKey, jsonToken, nil)
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("TokenGenerator.GenerateAccessToken"), err)
	}

	return accessToken, nil
}

func getJSONToken(now time.Time, ttl int, audience, issuer, subject string, claims map[string]string) paseto.JSONToken {
	token := paseto.JSONToken{
		Audience:   audience,
		Issuer:     issuer,
		Jti:        uuid.New().String(),
		Subject:    subject,
		Expiration: now.Add(time.Duration(ttl) * time.Minute),
		IssuedAt:   now,
		NotBefore:  now,
	}

	if claims != nil {
		for k, v := range claims {
			token.Set(k, v)
		}
	}

	return token
}

func (tg *pasetoTokenGenerator) GenerateRefreshToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("TokenGenerator.GenerateRefreshToken"), err)
	}

	return id.String(), nil
}

func NewGenerator(cfg config.TokenConfig) (Generator, error) {
	pem, err := base64.RawStdEncoding.DecodeString(cfg.EncodedSigningKey())
	if err != nil {
		return nil, liberr.WithArgs(liberr.Operation("TokenGenerator.NewTokenGenerator"), err)
	}

	privateKey, err := ssh.ParseRawPrivateKey(pem)
	if err != nil {
		return nil, liberr.WithArgs(liberr.Operation("TokenGenerator.NewTokenGenerator"), err)
	}

	ed25519PrivateKey, ok := privateKey.(*ed25519.PrivateKey)
	if !ok {
		return nil, liberr.WithArgs(liberr.Operation("TokenGenerator.NewTokenGenerator"), errors.New("invalid signing key"))
	}

	return &pasetoTokenGenerator{
		privateKey: *ed25519PrivateKey,
		audience:   cfg.Audience(),
		issuer:     cfg.Issuer(),
	}, nil
}
