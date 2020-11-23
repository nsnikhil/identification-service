package queue

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
)

type MockAMQP struct {
	mock.Mock
}

func (mock *MockAMQP) Push(name string, data []byte) error {
	args := mock.Called(name, data)
	return args.Error(0)
}

func (mock *MockAMQP) UnsafePush(name string, data []byte) error {
	args := mock.Called(name, data)
	return args.Error(0)
}

func (mock *MockAMQP) Stream(name string) (<-chan amqp.Delivery, error) {
	args := mock.Called(name)
	return args.Get(0).(<-chan amqp.Delivery), args.Error(1)
}

func (mock *MockAMQP) Close() error {
	args := mock.Called()
	return args.Error(0)
}
