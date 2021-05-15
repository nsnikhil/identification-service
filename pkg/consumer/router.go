package consumer

import (
	"fmt"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
	"identification-service/pkg/session"
)

type MessageRouter interface {
	Route(queueName string, message []byte) error
}

type ampqMessageRouter struct {
	cfg config.QueueConfig
	ss  session.Service
}

func (amr *ampqMessageRouter) Route(queueName string, message []byte) error {
	wrap := func(err error) error {
		return erx.WithArgs(erx.Operation("router.route"), err)
	}

	handler, err := getHandler(queueName, amr)
	if err != nil {
		return wrap(err)
	}

	if err := handler.Handle(message); err != nil {
		return wrap(err)
	}

	return nil
}

func getHandler(topic string, amr *ampqMessageRouter) (MessageHandler, error) {
	switch topic {
	case amr.cfg.UpdatePasswordQueueName():
		return NewUpdatePasswordHandler(amr.ss), nil
	default:
		return nil, fmt.Errorf("no handler found for the topics %s", topic)
	}
}

func NewMessageRouter(cfg config.QueueConfig, ss session.Service) MessageRouter {
	return &ampqMessageRouter{
		cfg: cfg,
		ss:  ss,
	}
}
