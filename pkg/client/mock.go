package client

import "github.com/stretchr/testify/mock"

type MockService struct {
	mock.Mock
}

func (mock *MockService) CreateClient(name string, accessTokenTTL int, sessionTTL int) (string, error) {
	args := mock.Called(name, accessTokenTTL, sessionTTL)
	return args.String(0), args.Error(1)
}

func (mock *MockService) RevokeClient(id string) error {
	args := mock.Called(id)
	return args.Error(0)
}

func (mock *MockService) GetClientTTL(name, secret string) (int, int, error) {
	args := mock.Called(name, secret)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (mock *MockService) ValidateClientCredentials(name, secret string) error {
	args := mock.Called(name, secret)
	return args.Error(0)
}
