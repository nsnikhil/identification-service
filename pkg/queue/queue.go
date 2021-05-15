package queue

import (
	"github.com/nsnikhil/erx"
	"github.com/streadway/amqp"
)

type Queue interface {
	Push(queueName string, data []byte) error
	Consume(queueName string) (<-chan amqp.Delivery, error)
	Close() error
}

type ampqQueue struct {
	ch *amqp.Channel
}

func (aq *ampqQueue) Push(queueName string, data []byte) error {
	err := aq.declareQueue(queueName)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{ContentType: "text/plain", Body: data}

	err = aq.ch.Publish("", queueName, false, false, msg)
	if err != nil {
		return erx.WithArgs(erx.WithArgs("Queue.Push"), err)
	}

	return nil
}

func (aq *ampqQueue) Consume(queueName string) (<-chan amqp.Delivery, error) {
	err := aq.declareQueue(queueName)
	if err != nil {
		return nil, err
	}

	dc, err := aq.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, erx.WithArgs(erx.WithArgs("Queue.Consume"), err)
	}

	return dc, nil
}

func (aq *ampqQueue) Close() error {
	err := aq.ch.Close()
	if err != nil {
		return erx.WithArgs(erx.WithArgs("Queue.Close"), err)
	}

	return nil
}

func (aq *ampqQueue) declareQueue(queueName string) error {
	_, err := aq.ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		return erx.WithArgs(erx.WithArgs("Queue.declareQueue"), err)
	}

	return nil
}

func NewQueue(ch *amqp.Channel) Queue {
	return &ampqQueue{
		ch: ch,
	}
}
