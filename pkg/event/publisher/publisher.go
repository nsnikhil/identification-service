package publisher

import (
	"errors"
	"fmt"
	"identification-service/pkg/event"
	"identification-service/pkg/liberr"
	"identification-service/pkg/queue"
)

type Publisher interface {
	Publish(eventCode event.Code, data interface{}) error
}

type eventPublisher struct {
	queue        queue.AMQP
	queueCodeMap map[event.Code]string
}

func (ep *eventPublisher) Publish(code event.Code, data interface{}) error {
	wrap := func(err error) error { return liberr.WithOp("Publisher.Publish", err) }

	queueName, ok := ep.queueCodeMap[code]
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

func NewPublisher(queue queue.AMQP, queueMap map[string]string) (Publisher, error) {
	queueCodeMap, err := parseQueueMap(queueMap)
	if err != nil {
		return nil, liberr.WithOp("Publisher.NewPublisher", err)
	}

	return &eventPublisher{
		queue:        queue,
		queueCodeMap: queueCodeMap,
	}, nil
}

//TODO: REFACTOR
func parseQueueMap(queueMap map[string]string) (map[event.Code]string, error) {
	if len(queueMap) == 0 {
		return nil, errors.New("queue map is empty")
	}

	res := make(map[event.Code]string)

	for eventCode, queueName := range queueMap {

		ok := event.CodeMap[event.Code(eventCode)]
		if !ok {
			return nil, fmt.Errorf("invalid event code %s", eventCode)
		}

		res[event.Code(eventCode)] = queueName
	}

	return res, nil
}
