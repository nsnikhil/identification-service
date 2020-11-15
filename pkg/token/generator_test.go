package token_test

import (
	"crypto"
	"crypto/ed25519"
	"errors"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/config"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/liberr"
	"identification-service/pkg/token"
	"regexp"
	"testing"
)

const userID = "86d690dd-92a0-40ac-ad48-110c951e3cb8"

var cfg = config.NewConfig("../../local.env").TokenConfig()

var pubKey = ed25519.PublicKey{6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}
var priKey = ed25519.PrivateKey{3, 195, 208, 247, 190, 104, 63, 62, 164, 50, 63, 217, 229, 215, 179, 62, 223, 104, 197, 43, 164, 164, 231, 1, 22, 70, 154, 130, 109, 98, 88, 210, 6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}

func TestCreateNewGeneratorSuccess(t *testing.T) {
	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, nil)

	_, err := token.NewGenerator(cfg, mockGenerator)
	require.NoError(t, err)
}

func TestCreateNewGeneratorFailure(t *testing.T) {
	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, liberr.WithArgs(errors.New("failed to genrate")))

	_, err := token.NewGenerator(cfg, mockGenerator)
	require.Error(t, err)
}

func TestAuthTokenGenerateAccessToken(t *testing.T) {
	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, priKey, nil)

	generator, err := token.NewGenerator(cfg, mockGenerator)
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

	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, priKey, nil)

	generator, err := token.NewGenerator(cfg, mockGenerator)
	require.NoError(t, err)

	refreshToken, err := generator.GenerateRefreshToken()
	require.NoError(t, err)

	assert.True(t, isValidUUID(refreshToken))
}

func validateTokens(t *testing.T, accessToken string) {
	var payload paseto.JSONToken

	_, err := paseto.Parse(accessToken, &payload, nil, nil, map[paseto.Version]crypto.PublicKey{paseto.Version2: pubKey})
	require.NoError(t, err)

	assert.Equal(t, "identification-service", payload.Issuer)
}
