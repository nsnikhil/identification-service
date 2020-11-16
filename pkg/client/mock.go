package client

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (mock *MockService) CreateClient(ctx context.Context, name string, accessTokenTTL, sessionTTL, maxActiveSessions int) (string, string, error) {
	args := mock.Called(ctx, name, accessTokenTTL, sessionTTL, maxActiveSessions)
	return args.String(0), args.String(1), args.Error(2)
}

func (mock *MockService) RevokeClient(ctx context.Context, id string) error {
	args := mock.Called(ctx, id)
	return args.Error(0)
}

func (mock *MockService) GetClient(ctx context.Context, name, secret string) (Client, error) {
	args := mock.Called(ctx, name, secret)
	return args.Get(0).(Client), args.Error(1)
}

type MockStore struct {
	mock.Mock
}

func (mock *MockStore) RevokeClient(ctx context.Context, id string) (int64, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (mock *MockStore) CreateClient(ctx context.Context, client Client) (string, error) {
	args := mock.Called(ctx, client)
	return args.String(0), args.Error(1)
}

func (mock *MockStore) GetClient(ctx context.Context, name, secret string) (Client, error) {
	args := mock.Called(ctx, name, secret)
	return args.Get(0).(Client), args.Error(1)
}
