package internal

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (mock *MockStore) CreateUser(ctx context.Context, user User) (string, error) {
	args := mock.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (mock *MockStore) GetUser(ctx context.Context, email string) (User, error) {
	args := mock.Called(ctx, email)
	return args.Get(0).(User), args.Error(1)
}

func (mock *MockStore) UpdatePassword(ctx context.Context, userID string, newPasswordHash string, newPasswordSalt []byte) (int64, error) {
	args := mock.Called(ctx, userID, newPasswordHash, newPasswordSalt)
	return args.Get(0).(int64), args.Error(1)
}
