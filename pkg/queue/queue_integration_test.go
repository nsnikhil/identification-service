package queue_test

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"testing"
	"time"
)

type queueTestSuite struct {
	qu queue.Queue
	ch *amqp.Channel
	suite.Suite
}

func (qt *queueTestSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env").QueueConfig()

	ch, err := queue.NewHandler(cfg).GetChannel()
	qt.Require().NoError(err)

	time.Sleep(time.Second)

	qt.ch = ch
	qt.qu = queue.NewQueue(ch)
}

func (qt *queueTestSuite) TearDownSuite() {
	qt.Require().NoError(qt.qu.Close())
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, new(queueTestSuite))
}

func (qt *queueTestSuite) TestQueuePush() {
	queueName := test.RandString(8)
	data := test.RandBytes(8)

	defer deleteQueue(qt, queueName)

	done := make(chan bool)

	go func() {
		err := qt.qu.Push(queueName, data)
		qt.Require().NoError(err)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 10):
		qt.Fail("pushed failed, timed out")
	}
}

func (qt *queueTestSuite) TestQueueConsume() {
	queueName := test.RandString(8)
	data := test.RandBytes(8)

	defer deleteQueue(qt, queueName)

	go func() {
		err := qt.qu.Push(queueName, data)
		qt.Require().NoError(err)
	}()

	value := make(chan []byte)

	go func() {
		ch, err := qt.qu.Consume(queueName)
		qt.Require().NoError(err)
		msg := <-ch
		value <- msg.Body
	}()

	select {
	case v := <-value:
		qt.Assert().Equal(data, v)
	case <-time.After(time.Second * 10):
		qt.Fail("consume failed, timed out")
	}
}

func deleteQueue(qt *queueTestSuite, queueName string) {
	_, err := qt.ch.QueueDelete(queueName, false, false, true)
	qt.Require().NoError(err)
}
