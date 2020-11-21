package session

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (mock *MockService) LoginUser(ctx context.Context, email, password string) (string, string, error) {
	args := mock.Called(ctx, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (mock *MockService) LogoutUser(ctx context.Context, refreshToken string) error {
	args := mock.Called(ctx, refreshToken)
	return args.Error(0)
}

func (mock *MockService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := mock.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

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

func (mock *MockStore) GetActiveSessionsCount(ctx context.Context, userID string) (int, error) {
	args := mock.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (mock *MockStore) RevokeSessions(ctx context.Context, refreshTokens ...string) (int64, error) {
	args := mock.Called(ctx, refreshTokens)
	return args.Get(0).(int64), args.Error(1)
}

func (mock *MockStore) RevokeLastNSessions(ctx context.Context, userID string, n int) (int64, error) {
	args := mock.Called(ctx, userID, n)
	return args.Get(0).(int64), args.Error(1)
}
