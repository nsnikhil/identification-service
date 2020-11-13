package internal

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (mock *MockStore) CreateSession(ctx context.Context, session Session) (string, error) {
	args := mock.Called(ctx, session)
	return args.String(0), args.Error(1)
}

func (mock *MockStore) GetSession(ctx context.Context, refreshToken string) (Session, error) {
	args := mock.Called(ctx, refreshToken)
	return args.Get(0).(Session), args.Error(1)
}

func (mock *MockStore) RevokeSession(ctx context.Context, refreshToken string) (int64, error) {
	args := mock.Called(ctx, refreshToken)
	return args.Get(0).(int64), args.Error(1)
}
