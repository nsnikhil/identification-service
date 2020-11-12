package internal

import "github.com/stretchr/testify/mock"

type MockStore struct {
	mock.Mock
}

func (mock *MockStore) RevokeClient(id string) (int64, error) {
	args := mock.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func (mock *MockStore) CreateClient(client Client) (string, error) {
	args := mock.Called(client)
	return args.String(0), args.Error(1)
}

func (mock *MockStore) GetClient(name, secret string) (Client, error) {
	args := mock.Called(name, secret)
	return args.Get(0).(Client), args.Error(1)
}
