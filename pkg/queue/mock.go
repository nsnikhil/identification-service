package queue

import "github.com/stretchr/testify/mock"

type MockQueue struct {
	mock.Mock
}

func (mock *MockQueue) Push(data []byte) error {
	args := mock.Called(data)
	return args.Error(0)
}

func (mock *MockQueue) UnsafePush(data []byte) error {
	args := mock.Called(data)
	return args.Error(0)
}

func (mock *MockQueue) Close() error {
	args := mock.Called()
	return args.Error(0)
}
