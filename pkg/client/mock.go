package client

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (mock *MockService) CreateClient(ctx context.Context, name string, accessTokenTTL int, sessionTTL int) (string, error) {
	args := mock.Called(ctx, name, accessTokenTTL, sessionTTL)
	return args.String(0), args.Error(1)
}

func (mock *MockService) RevokeClient(ctx context.Context, id string) error {
	args := mock.Called(ctx, id)
	return args.Error(0)
}

func (mock *MockService) GetClientTTL(ctx context.Context, name, secret string) (int, int, error) {
	args := mock.Called(ctx, name, secret)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (mock *MockService) ValidateClientCredentials(ctx context.Context, name, secret string) error {
	args := mock.Called(ctx, name, secret)
	return args.Error(0)
}
