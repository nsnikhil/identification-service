package session

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (mock *MockService) LoginUser(ctx context.Context, clientName, clientSecret, email, password string) (string, string, error) {
	args := mock.Called(ctx, clientName, clientSecret, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (mock *MockService) LogoutUser(ctx context.Context, refreshToken string) error {
	args := mock.Called(ctx, refreshToken)
	return args.Error(0)
}

func (mock *MockService) RefreshToken(ctx context.Context, clientName, clientSecret, refreshToken string) (string, error) {
	args := mock.Called(ctx, clientName, clientSecret, refreshToken)
	return args.String(0), args.Error(1)
}
