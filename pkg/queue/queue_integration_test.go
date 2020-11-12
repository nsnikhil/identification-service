// build integration_test

package queue_test

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"identification-service/pkg/config"
	"identification-service/pkg/queue"
	"testing"
	"time"
)

//TODO: ADD TEST FOR FAILURE SCENARIOS

var cfg = config.NewConfig("../../local.env").AMPQConfig()

type queueTestSuite struct {
	qu queue.Queue
	suite.Suite
}

func (qt *queueTestSuite) SetupSuite() {
	qt.qu = getQueue()
}

func (qt *queueTestSuite) AfterTest(suiteName, testName string) {
	purgeMessages(qt.T())
}

func (qt *queueTestSuite) TestPushSuccess() {
	data := []byte("test message")

	err := qt.qu.Push(data)
	require.NoError(qt.T(), err)
}

func (qt *queueTestSuite) TestUnSafePushSuccess() {
	data := []byte("test message")

	err := qt.qu.UnsafePush(data)
	require.NoError(qt.T(), err)
}

func TestQueue(t *testing.T) {
	suite.Run(t, new(queueTestSuite))
}

func purgeMessages(t *testing.T) {
	conn, err := amqp.Dial(cfg.Address())
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	ch, err := conn.Channel()
	require.NoError(t, err)

	_, err = ch.QueuePurge(cfg.QueueName(), true)
	require.NoError(t, err)
}

func getQueue() queue.Queue {
	qu := queue.NewQueue(cfg.QueueName(), cfg.Address(), zap.NewNop())

	time.Sleep(time.Millisecond * 100)

	return qu
}
