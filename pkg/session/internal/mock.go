package internal

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (mock *MockStore) CreateSession(session Session) (string, error) {
	args := mock.Called(session)
	return args.String(0), args.Error(1)
}

func (mock *MockStore) GetSession(refreshToken string) (Session, error) {
	args := mock.Called(refreshToken)
	return args.Get(0).(Session), args.Error(1)
}

func (mock *MockStore) RevokeSession(refreshToken string) (int64, error) {
	args := mock.Called(refreshToken)
	return args.Get(0).(int64), args.Error(1)
}
