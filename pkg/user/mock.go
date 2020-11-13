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
