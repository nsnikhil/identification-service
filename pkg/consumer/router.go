package consumer

import (
	"fmt"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
	"identification-service/pkg/session"
)

type MessageRouter interface {
	Route(topic string, message []byte) error
}

type kafkaMessageRouter struct {
	cfg config.KafkaConfig
	ss  session.Service
}

func (kmr *kafkaMessageRouter) Route(topic string, message []byte) error {
	wrap := func(err error) error {
		return erx.WithArgs(erx.Operation("router.route"), err)
	}

	handler, err := getHandler(topic, kmr)
	if err != nil {
		return wrap(err)
	}

	if err := handler.Handle(message); err != nil {
		return wrap(err)
	}

	return nil
}

func getHandler(topic string, kcr *kafkaMessageRouter) (MessageHandler, error) {
	switch topic {
	case kcr.cfg.UpdatePasswordTopicName():
		return NewUpdatePasswordHandler(kcr.ss), nil
	default:
		return nil, fmt.Errorf("no handler found for the topics %s", topic)
	}
}

func NewMessageRouter(cfg config.KafkaConfig, ss session.Service) MessageRouter {
	return &kafkaMessageRouter{
		cfg: cfg,
		ss:  ss,
	}
}
