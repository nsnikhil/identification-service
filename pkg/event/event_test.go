package event_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/event"
	"identification-service/pkg/test"
	"testing"
)

func TestCreateNewEventSuccess(t *testing.T) {
	_, err := event.NewEvent("sign-up", "some data")
	assert.Nil(t, err)
}

func TestCreateNewEventFailure(t *testing.T) {
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
			_, err := event.NewEvent(testCase.eventCode, testCase.data)
			assert.Error(t, err)
		})
	}

}

func TestEventToBytesSuccess(t *testing.T) {
	ev, err := event.NewEvent("sign-up", "some data")
	assert.Nil(t, err)

	_, err = ev.ToBytes()
	assert.Nil(t, err)
}

func TestToBytesFailureFailure(t *testing.T) {
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
}

func TestFromBytesSuccess(t *testing.T) {
	ev, err := event.NewEvent("sign-up", "some data")
	assert.Nil(t, err)

	b, err := ev.ToBytes()
	assert.Nil(t, err)

	bev, err := event.FromBytes(b)
	assert.Nil(t, err)

	assert.Equal(t, ev, bev)
}

func TestFromBytesFailure(t *testing.T) {
	_, err := event.FromBytes([]byte{1, 2, 3, 4})
	assert.Error(t, err)
}
