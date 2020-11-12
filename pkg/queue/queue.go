package queue

import (
	"errors"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"time"
)

//TODO: PICK THE CONST FROM CONFIG
const (
	reconnectDelay = 5 * time.Second
	reInitDelay    = 2 * time.Second
	resendDelay    = 5 * time.Second
)

type Queue interface {
	Push(data []byte) error
	UnsafePush(data []byte) error
	Close() error
}

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("session is shutting down")
)

//TODO: CHECK IF THE LOGGER IS NEEDED HERE ?
type rabbitMQ struct {
	name            string
	logger          *zap.Logger
	connection      *amqp.Connection
	channel         *amqp.Channel
	done            chan bool
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	isReady         bool
}

func NewQueue(name string, addr string, lgr *zap.Logger) Queue {
	mq := rabbitMQ{
		logger: lgr,
		name:   name,
		done:   make(chan bool),
	}
	go mq.handleReconnect(addr)
	return &mq
}

func (mq *rabbitMQ) handleReconnect(addr string) {
	for {
		mq.isReady = false
		mq.logger.Info("Attempting to connect")

		conn, err := mq.connect(addr)
		if err != nil {
			mq.logger.Warn("Failed to connect. Retrying...")

			select {
			case <-mq.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := mq.handleReInit(conn); done {
			break
		}
	}
}

func (mq *rabbitMQ) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	mq.changeConnection(conn)
	mq.logger.Info("Connected!")
	return conn, nil
}

func (mq *rabbitMQ) handleReInit(conn *amqp.Connection) bool {
	for {
		mq.isReady = false

		err := mq.init(conn)
		if err != nil {
			mq.logger.Warn("Failed to initialize channel. Retrying...")

			select {
			case <-mq.done:
				return true
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-mq.done:
			return true
		case <-mq.notifyConnClose:
			mq.logger.Warn("Connection closed. Reconnecting...")
			return false
		case <-mq.notifyChanClose:
			mq.logger.Warn("Channel closed. Re-running init...")
		}
	}
}

func (mq *rabbitMQ) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(mq.name, false, false, false, false, nil)
	if err != nil {
		return err
	}

	mq.changeChannel(ch)
	mq.isReady = true
	mq.logger.Info("Setup!")

	return nil
}

func (mq *rabbitMQ) changeConnection(connection *amqp.Connection) {
	mq.connection = connection
	mq.notifyConnClose = make(chan *amqp.Error)
	mq.connection.NotifyClose(mq.notifyConnClose)
}

func (mq *rabbitMQ) changeChannel(channel *amqp.Channel) {
	mq.channel = channel
	mq.notifyChanClose = make(chan *amqp.Error)
	mq.notifyConfirm = make(chan amqp.Confirmation, 1)
	mq.channel.NotifyClose(mq.notifyChanClose)
	mq.channel.NotifyPublish(mq.notifyConfirm)
}

func (mq *rabbitMQ) Push(data []byte) error {
	if !mq.isReady {
		return errors.New("failed to push push: not connected")
	}

	for {
		err := mq.UnsafePush(data)
		if err != nil {
			mq.logger.Warn("Push failed. Retrying...")
			select {
			case <-mq.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}

		select {
		case confirm := <-mq.notifyConfirm:
			if confirm.Ack {
				mq.logger.Info("Push confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}

		mq.logger.Warn("Push didn't confirm. Retrying...")
	}
}

func (mq *rabbitMQ) UnsafePush(data []byte) error {
	if !mq.isReady {
		return errNotConnected
	}

	return mq.channel.Publish("", mq.name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
	})
}

func (mq *rabbitMQ) Stream() (<-chan amqp.Delivery, error) {
	if !mq.isReady {
		return nil, errNotConnected
	}

	return mq.channel.Consume(mq.name, "", false, false, false, false, nil)
}

func (mq *rabbitMQ) Close() error {
	if !mq.isReady {
		return errAlreadyClosed
	}

	err := mq.channel.Close()
	if err != nil {
		return err
	}

	err = mq.connection.Close()
	if err != nil {
		return err
	}

	close(mq.done)
	mq.isReady = false
	return nil
}
