package internal

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (mock *MockStore) CreateUser(user User) (string, error) {
	args := mock.Called(user)
	return args.String(0), args.Error(1)
}

func (mock *MockStore) GetUser(email string) (User, error) {
	args := mock.Called(email)
	return args.Get(0).(User), args.Error(1)
}

func (mock *MockStore) UpdatePassword(userID string, newPasswordHash string, newPasswordSalt []byte) (int64, error) {
	args := mock.Called(userID, newPasswordHash, newPasswordSalt)
	return args.Get(0).(int64), args.Error(1)
}
