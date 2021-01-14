package publisher_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/event"
	"identification-service/pkg/event/publisher"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"testing"
)

func TestCreatePublisherSuccess(t *testing.T) {
	queueMap := map[string]string{"sign-up": "id-sign-up"}

	_, err := publisher.NewPublisher(&queue.MockAMQP{}, queueMap)
	assert.Nil(t, err)
}

func TestCreatePublisherFailure(t *testing.T) {
	_, err := publisher.NewPublisher(&queue.MockAMQP{}, map[string]string{})
	assert.Error(t, err)

	queueMap := map[string]string{"other": "id-other"}
	_, err = publisher.NewPublisher(&queue.MockAMQP{}, queueMap)
	assert.Error(t, err)
}

func TestPublisherPublishSuccess(t *testing.T) {
	mockQueue := &queue.MockAMQP{}
	mockQueue.On("UnsafePush", "id-sign-up", mock.AnythingOfType("[]uint8")).Return(nil)

	queueMap := map[string]string{"sign-up": "id-sign-up"}

	pb, err := publisher.NewPublisher(mockQueue, queueMap)
	require.NoError(t, err)

	err = pb.Publish(event.SignUp, test.NewUUID())
	assert.Nil(t, err)
}

func TestPublishFailureWhenEventCreationFails(t *testing.T) {
	mockQueue := &queue.MockAMQP{}
	mockQueue.On("UnsafePush", "id-sign-up", mock.AnythingOfType("[]uint8")).Return(nil)

	queueMap := map[string]string{"sign-up": "id-sign-up"}

	pb, err := publisher.NewPublisher(mockQueue, queueMap)
	require.NoError(t, err)

	testCases := map[string]struct {
		eventCode event.Code
		data      interface{}
	}{
		"test failure when event code is invalid": {
			eventCode: event.Code("other"),
			data:      "some data",
		},
		"test failure when data is nil": {
			eventCode: event.SignUp,
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
			err = pb.Publish(testCase.eventCode, testCase.data)
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

	pb, err := publisher.NewPublisher(mockQueue, queueMap)
	require.NoError(t, err)

	err = pb.Publish(event.SignUp, test.NewUUID())
	assert.Error(t, err)
}
