package consumer

import "github.com/stretchr/testify/mock"

type MockConsumer struct {
	mock.Mock
}

func (mock *MockConsumer) Start() {
	mock.Called()
}

func (mock *MockConsumer) Close() error {
	args := mock.Called()
	return args.Error(0)
}
