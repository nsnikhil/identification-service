package queue_test

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"testing"
	"time"
)

//TODO: ADD TEST FOR FAILURE SCENARIOS
var cfg = config.NewConfig("../../local.env")

type queueTestSuite struct {
	qu queue.AMQP
	suite.Suite
}

func (qt *queueTestSuite) SetupSuite() {
	qt.qu = getQueue()
}

func (qt *queueTestSuite) TearDownSuite() {
	require.NoError(qt.T(), qt.qu.Close())
}

func (qt *queueTestSuite) AfterTest(suiteName, testName string) {
	purgeMessages(qt.T())
}

func (qt *queueTestSuite) TestPushSuccess() {
	data := []byte("test message")

	err := qt.qu.Push(test.QueueName, data)
	require.NoError(qt.T(), err)
}

func (qt *queueTestSuite) TestUnSafePushSuccess() {
	data := []byte("test message")

	err := qt.qu.UnsafePush(test.QueueName, data)
	require.NoError(qt.T(), err)
}

func TestQueue(t *testing.T) {
	suite.Run(t, new(queueTestSuite))
}

func purgeMessages(t *testing.T) {
	conn, err := amqp.Dial(cfg.AMPQConfig().Address())
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	ch, err := conn.Channel()
	require.NoError(t, err)

	for _, queueName := range cfg.EventConfig().QueueMap() {
		_, err = ch.QueuePurge(queueName, true)
		require.NoError(t, err)
	}
}

func getQueue() queue.AMQP {
	qu := queue.NewAMQP(cfg.AMPQConfig().Address())

	time.Sleep(time.Millisecond * 100)

	return qu
}
