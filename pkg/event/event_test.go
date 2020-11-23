package event_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/event"
	"testing"
)

func TestCreateNewEventSuccess(t *testing.T) {
	_, err := event.NewEvent(event.SignUp, "some data")
	assert.Nil(t, err)
}

func TestCreateNewEventFailure(t *testing.T) {
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
			_, err := event.NewEvent(testCase.eventCode, testCase.data)
			assert.Error(t, err)
		})
	}

}

func TestEventToBytesSuccess(t *testing.T) {
	ev, err := event.NewEvent(event.SignUp, "some data")
	assert.Nil(t, err)

	_, err = ev.ToBytes()
	assert.Nil(t, err)
}

func TestToBytesFailureFailure(t *testing.T) {
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
}

func TestFromBytes(t *testing.T) {
	ev, err := event.NewEvent(event.SignUp, "some data")
	assert.Nil(t, err)

	b, err := ev.ToBytes()
	assert.Nil(t, err)

	bev, err := event.FromBytes(b)
	assert.Nil(t, err)

	assert.Equal(t, ev, bev)
}
