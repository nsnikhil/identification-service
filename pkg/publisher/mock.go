package publisher

import (
	"github.com/stretchr/testify/mock"
)

type MockPublisher struct {
	mock.Mock
}

func (mock *MockPublisher) Publish(eventCode string, data interface{}) error {
	args := mock.Called(eventCode, data)
	return args.Error(0)
}
