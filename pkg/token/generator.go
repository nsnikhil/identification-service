package token

import (
	"crypto"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"identification-service/pkg/config"
	"identification-service/pkg/libcrypto"
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
	privateKey crypto.PrivateKey
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

func NewGenerator(cfg config.TokenConfig, keyGenerator libcrypto.Ed25519Generator) (Generator, error) {
	_, priKey, err := keyGenerator.FromEncodedPem(cfg.EncodedSigningKey())
	if err != nil {
		return nil, liberr.WithArgs(liberr.Operation("TokenGenerator.NewTokenGenerator"), err)
	}

	return &pasetoTokenGenerator{
		privateKey: priKey,
		audience:   cfg.Audience(),
		issuer:     cfg.Issuer(),
	}, nil
}
