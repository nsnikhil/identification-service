package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"identification-service/pkg/liberr"
)

type Event struct {
	Code Code        `json:"code"`
	Data interface{} `json:"data"`
}

func (e Event) ToBytes() ([]byte, error) {
	err := validate(e.Code, e.Data)
	if err != nil {
		return nil, liberr.WithArgs(liberr.Operation("Event.ToBytes"), liberr.ValidationError, err)
	}

	b, err := json.Marshal(e)
	if err != nil {
		return nil, liberr.WithOp("Event.toBytes", err)
	}

	return b, nil
}

func FromBytes(b []byte) (Event, error) {
	var e Event

	err := json.Unmarshal(b, &e)
	if err != nil {
		return Event{}, liberr.WithOp("Event.fromBytes", err)
	}

	return e, nil
}

func NewEvent(code Code, data interface{}) (Event, error) {
	err := validate(code, data)
	if err != nil {
		return Event{}, liberr.WithArgs(liberr.Operation("Event.NewEvent"), liberr.ValidationError, err)
	}

	return Event{Code: code, Data: data}, nil
}

func validate(code Code, data interface{}) error {
	if ok := CodeMap[code]; !ok {
		return fmt.Errorf("invalid error code %v", code)
	}

	if data == nil {
		return errors.New("data cannot be nil")
	}

	return nil
}
