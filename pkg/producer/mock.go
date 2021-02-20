package producer

import "github.com/stretchr/testify/mock"

type MockProducer struct {
	mock.Mock
}

func (mock *MockProducer) Produce(topic string, value []byte) (int32, int64, error) {
	args := mock.Called(topic, value)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (mock *MockProducer) Close() error {
	args := mock.Called()
	return args.Error(0)
}
