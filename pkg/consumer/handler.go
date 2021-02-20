package consumer

import (
	"context"
	"identification-service/pkg/liberr"
	"identification-service/pkg/session"
)

type MessageHandler interface {
	Handle(msg []byte) error
}

type updatePasswordHandler struct {
	ss session.Service
}

func (uph *updatePasswordHandler) Handle(msg []byte) error {
	if err := uph.ss.RevokeAllSessions(context.Background(), string(msg)); err != nil {
		return liberr.WithOp("updatePasswordHandler", err)
	}

	return nil
}

func NewUpdatePasswordHandler(ss session.Service) MessageHandler {
	return &updatePasswordHandler{
		ss: ss,
	}
}
