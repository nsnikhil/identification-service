package publisher

import (
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/event"
)

type MockPublisher struct {
	mock.Mock
}

func (mock *MockPublisher) Publish(eventCode event.Code, data interface{}) error {
	args := mock.Called(eventCode, data)
	return args.Error(0)
}
