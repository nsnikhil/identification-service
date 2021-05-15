package queue

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
)

type MockQueue struct {
	mock.Mock
}

func (mock *MockQueue) Push(queueName string, data []byte) error {
	args := mock.Called(queueName, data)
	return args.Error(0)
}

func (mock *MockQueue) Consume(queueName string) (<-chan amqp.Delivery, error) {
	args := mock.Called(queueName)
	return args.Get(0).(chan amqp.Delivery), args.Error(1)
}

func (mock *MockQueue) Close() error {
	args := mock.Called()
	return args.Error(0)
}
