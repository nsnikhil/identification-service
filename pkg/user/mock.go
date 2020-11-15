package user

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (mock *MockService) CreateUser(ctx context.Context, name, email, password string) (string, error) {
	args := mock.Called(ctx, name, email, password)
	return args.String(0), args.Error(1)
}

func (mock *MockService) UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	args := mock.Called(ctx, email, oldPassword, newPassword)
	return args.Error(0)
}

func (mock *MockService) GetUserID(ctx context.Context, email, password string) (string, error) {
	args := mock.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

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
