package token_test

import (
	"crypto"
	"crypto/ed25519"
	"encoding/base64"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"identification-service/pkg/config"
	"identification-service/pkg/token"
	"regexp"
	"testing"
)

const userID = "86d690dd-92a0-40ac-ad48-110c951e3cb8"

var cfg = config.NewConfig("../../local.env").TokenConfig()

func TestAuthTokenGenerateAccessToken(t *testing.T) {
	generator, err := token.NewGenerator(cfg)
	require.NoError(t, err)

	accessToken, err := generator.GenerateAccessToken(10, userID, nil)
	require.NoError(t, err)

	validateTokens(t, accessToken)
}

func TestAuthTokenGenerateRefreshToken(t *testing.T) {
	isValidUUID := func(uuid string) bool {
		r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
		return r.MatchString(uuid)
	}

	generator, err := token.NewGenerator(cfg)
	require.NoError(t, err)

	refreshToken, err := generator.GenerateRefreshToken()
	require.NoError(t, err)

	assert.True(t, isValidUUID(refreshToken))
}

func validateTokens(t *testing.T, accessToken string) {
	pem, err := base64.RawStdEncoding.DecodeString(cfg.EncodedSigningKey())
	require.NoError(t, err)

	privateKey, err := ssh.ParseRawPrivateKey(pem)
	require.NoError(t, err)

	publicKey := privateKey.(*ed25519.PrivateKey).Public()

	var payload paseto.JSONToken

	_, err = paseto.Parse(accessToken, &payload, nil, pem, map[paseto.Version]crypto.PublicKey{paseto.Version2: publicKey})
	require.NoError(t, err)

	assert.Equal(t, "identification-service", payload.Issuer)
}
