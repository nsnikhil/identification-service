package session

import "github.com/stretchr/testify/mock"

type MockService struct {
	mock.Mock
}

func (mock *MockService) LoginUser(clientName, clientSecret, email, password string) (string, string, error) {
	args := mock.Called(clientName, clientSecret, email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (mock *MockService) LogoutUser(refreshToken string) error {
	args := mock.Called(refreshToken)
	return args.Error(0)
}

func (mock *MockService) RefreshToken(clientName, clientSecret, refreshToken string) (string, error) {
	args := mock.Called(clientName, clientSecret, refreshToken)
	return args.String(0), args.Error(1)
}
