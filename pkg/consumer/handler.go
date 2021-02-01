package consumer

import (
	"context"
	"errors"
	"fmt"
	"identification-service/pkg/event"
	"identification-service/pkg/liberr"
	"identification-service/pkg/session"
)

func handleUpdatePassword(ss session.Service, event event.Event) error {
	wrap := func(err error) error { return liberr.WithOp("handleUpdatePassword", err) }

	data := event.Data
	if data == nil {
		return wrap(errors.New("data is nil"))
	}

	userID, ok := data.(string)
	if !ok {
		return wrap(fmt.Errorf("invalid data %v", data))
	}

	err := ss.RevokeAllSessions(context.Background(), userID)
	if err != nil {
		return wrap(err)
	}

	return nil
}
