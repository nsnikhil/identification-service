package queue

import (
	"github.com/nsnikhil/erx"
	"github.com/streadway/amqp"
	"identification-service/pkg/config"
)

type Handler interface {
	GetChannel() (*amqp.Channel, error)
}

type ampqChannelHandler struct{ address string }

func (ach *ampqChannelHandler) GetChannel() (*amqp.Channel, error) {
	wrap := func(err error) error {
		return erx.WithArgs(erx.Operation("Handler.GetChannel"), err)
	}

	conn, err := amqp.Dial(ach.address)
	if err != nil {
		return nil, wrap(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func NewHandler(cfg config.QueueConfig) Handler {
	return &ampqChannelHandler{
		address: cfg.Address(),
	}
}
