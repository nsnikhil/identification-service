package publisher_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/event"
	"identification-service/pkg/publisher"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"testing"
)

func TestCreatePublisherSuccess(t *testing.T) {
	queueMap := map[string]string{"sign-up": "id-sign-up"}

	pb := publisher.NewPublisher(&queue.MockAMQP{}, queueMap)
	assert.NotNil(t, pb)
}

func TestPublisherPublishSuccess(t *testing.T) {
	mockQueue := &queue.MockAMQP{}
	mockQueue.On("UnsafePush", "id-sign-up", mock.AnythingOfType("[]uint8")).Return(nil)

	queueMap := map[string]string{"sign-up": "id-sign-up"}

	pb := publisher.NewPublisher(mockQueue, queueMap)

	err := pb.Publish("sign-up", test.NewUUID())
	assert.Nil(t, err)
}

func TestPublishFailureWhenEventCreationFails(t *testing.T) {
	mockQueue := &queue.MockAMQP{}
	mockQueue.On("UnsafePush", "id-sign-up", mock.AnythingOfType("[]uint8")).Return(nil)

	queueMap := map[string]string{"sign-up": "id-sign-up"}

	pb := publisher.NewPublisher(&queue.MockAMQP{}, queueMap)

	testCases := map[string]struct {
		eventCode string
		data      interface{}
	}{
		"test failure when event code is empty": {
			eventCode: test.EmptyString,
			data:      "some data",
		},
		"test failure when data is nil": {
			eventCode: "sign-up",
			data:      nil,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ev := event.Event{Code: testCase.eventCode, Data: testCase.data}
			_, err := ev.ToBytes()
			assert.Error(t, err)
		})
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			err := pb.Publish(testCase.eventCode, testCase.data)
			assert.Error(t, err)
		})
	}
}

func TestPublishFailureWhenPushFails(t *testing.T) {
	mockQueue := &queue.MockAMQP{}
	mockQueue.On(
		"UnsafePush",
		"id-sign-up", mock.AnythingOfType("[]uint8"),
	).Return(errors.New("failed to push event"))

	queueMap := map[string]string{"sign-up": "id-sign-up"}

	pb := publisher.NewPublisher(mockQueue, queueMap)

	err := pb.Publish("sign-up", test.NewUUID())
	assert.Error(t, err)
}
