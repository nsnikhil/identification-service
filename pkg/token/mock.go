package token

import "github.com/stretchr/testify/mock"

type MockGenerator struct {
	mock.Mock
}

func (mock *MockGenerator) GenerateAccessToken(ttl int, subject string, claims map[string]string) (string, error) {
	args := mock.Called(ttl, subject, claims)
	return args.String(0), args.Error(1)
}

func (mock *MockGenerator) GenerateRefreshToken() (string, error) {
	args := mock.Called()
	return args.String(0), args.Error(1)
}
