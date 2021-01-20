package config

import "github.com/stretchr/testify/mock"

type TokenConfig interface {
	Audience() string
	Issuer() string
	EncodedSigningKey() string
}

type appTokenConfig struct {
	audience          string
	issuer            string
	encodedSigningKey string
}

func newTokenConfig() TokenConfig {
	return appTokenConfig{
		audience:          getString("TOKEN_AUDIENCE"),
		issuer:            getString("TOKEN_ISSUER"),
		encodedSigningKey: getString("ENCODED_TOKEN_SIGNING_KEY"),
	}
}

func (tc appTokenConfig) Audience() string {
	return tc.audience
}

func (tc appTokenConfig) Issuer() string {
	return tc.issuer
}

func (tc appTokenConfig) EncodedSigningKey() string {
	return tc.encodedSigningKey
}

type MockTokenConfig struct {
	mock.Mock
}

func (mock *MockTokenConfig) Audience() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockTokenConfig) Issuer() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockTokenConfig) EncodedSigningKey() string {
	args := mock.Called()
	return args.String(0)
}
