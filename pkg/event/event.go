package event

import (
	"encoding/json"
	"errors"
	"identification-service/pkg/liberr"
)

type Event struct {
	Code string      `json:"code"`
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

func NewEvent(code string, data interface{}) (Event, error) {
	if err := validate(code, data); err != nil {
		return Event{}, liberr.WithArgs(liberr.Operation("Event.NewEvent"), liberr.ValidationError, err)
	}

	return Event{Code: code, Data: data}, nil
}

func validate(code string, data interface{}) error {
	if len(code) == 0 {
		return errors.New("code cannot be empty")
	}

	if data == nil {
		return errors.New("data cannot be nil")
	}

	return nil
}
