package user

import (
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (mock *MockService) CreateUser(name, email, password string) (string, error) {
	args := mock.Called(name, email, password)
	return args.String(0), args.Error(1)
}

func (mock *MockService) UpdatePassword(email, oldPassword, newPassword string) error {
	args := mock.Called(email, oldPassword, newPassword)
	return args.Error(0)
}

func (mock *MockService) GetUserID(email, password string) (string, error) {
	args := mock.Called(email, password)
	return args.String(0), args.Error(1)
}
