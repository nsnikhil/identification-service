package token_test

import (
	"crypto"
	"crypto/ed25519"
	"errors"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/test"
	"identification-service/pkg/token"
	"regexp"
	"testing"
)

type generatorTest struct {
	suite.Suite
	cfg config.TokenConfig
}

func (gt *generatorTest) SetupSuite() {
	mockClientConfig := &config.MockTokenConfig{}
	mockClientConfig.On("Audience").Return("user")
	mockClientConfig.On("Issuer").Return("identification-service")
	mockClientConfig.On("EncodedSigningKey").Return("signing-key")

	gt.cfg = mockClientConfig
}

func TestGenerator(t *testing.T) {
	suite.Run(t, new(generatorTest))
}

func (gt *generatorTest) TestCreateNewGeneratorSuccess() {
	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, nil)

	_, err := token.NewGenerator(gt.cfg, mockGenerator)
	gt.Require().NoError(err)
}

func (gt *generatorTest) TestCreateNewGeneratorFailure() {
	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, errors.New("failed to genrate"))

	_, err := token.NewGenerator(gt.cfg, mockGenerator)
	gt.Require().Error(err)
}

func (gt *generatorTest) TestAuthTokenGenerateAccessToken() {
	pub, pri := test.GenerateKey()

	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, pri, nil)

	generator, err := token.NewGenerator(gt.cfg, mockGenerator)
	gt.Require().NoError(err)

	accessToken, err := generator.GenerateAccessToken(10, test.NewUUID(), nil)
	gt.Require().NoError(err)

	var payload paseto.JSONToken

	_, err = paseto.Parse(accessToken, &payload, nil, nil, map[paseto.Version]crypto.PublicKey{paseto.Version2: pub})
	gt.Require().NoError(err)

	gt.Assert().Equal("identification-service", payload.Issuer)
}

func (gt *generatorTest) TestAuthTokenGenerateRefreshToken() {
	isValidUUID := func(uuid string) bool {
		r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
		return r.MatchString(uuid)
	}

	mockGenerator := &libcrypto.MockEd25519Generator{}
	mockGenerator.On(
		"FromEncodedPem",
		mock.AnythingOfType("string"),
	).Return(ed25519.PublicKey{}, test.ClientPriKey(), nil)

	generator, err := token.NewGenerator(gt.cfg, mockGenerator)
	gt.Require().NoError(err)

	refreshToken, err := generator.GenerateRefreshToken()
	gt.Require().NoError(err)

	gt.Assert().True(isValidUUID(refreshToken))
}
