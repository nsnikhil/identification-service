package publisher

import (
	"fmt"
	"identification-service/pkg/event"
	"identification-service/pkg/liberr"
	"identification-service/pkg/queue"
)

type Publisher interface {
	Publish(eventCode string, data interface{}) error
}

type eventPublisher struct {
	queue    queue.AMQP
	queueMap map[string]string
}

func (ep *eventPublisher) Publish(code string, data interface{}) error {
	wrap := func(err error) error { return liberr.WithOp("Publisher.Publish", err) }

	queueName, ok := ep.queueMap[code]
	if !ok {
		return wrap(fmt.Errorf("invalid event code %s", code))
	}

	ev, err := event.NewEvent(code, data)
	if err != nil {
		return wrap(err)
	}

	b, err := ev.ToBytes()
	if err != nil {
		return wrap(err)
	}

	err = ep.queue.UnsafePush(queueName, b)
	if err != nil {
		return wrap(err)
	}

	return nil
}

func NewPublisher(queue queue.AMQP, queueMap map[string]string) Publisher {
	return &eventPublisher{
		queue:    queue,
		queueMap: queueMap,
	}
}
