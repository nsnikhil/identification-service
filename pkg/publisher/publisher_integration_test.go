package publisher_test

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/publisher"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"testing"
	"time"
)

type publisherIntegrationTestSuite struct {
	qu  queue.AMQP
	pb  publisher.Publisher
	cfg config.EventConfig
	suite.Suite
}

func (pts *publisherIntegrationTestSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	qu := queue.NewAMQP(cfg.AMPQConfig().Address())
	time.Sleep(time.Millisecond * 100)

	pb := publisher.NewPublisher(qu, cfg.EventConfig().QueueMap())

	pts.qu = qu
	pts.pb = pb
	pts.cfg = cfg.EventConfig()
}

func (pts *publisherIntegrationTestSuite) TearDownSuite() {
	require.NoError(pts.T(), pts.qu.Close())
}

func (pts *publisherIntegrationTestSuite) AfterTest(suiteName, testName string) {
	purgeMessages(pts.T())
}

func (pts *publisherIntegrationTestSuite) TestPublisherPublishSuccess() {
	err := pts.pb.Publish(pts.cfg.SignUpEventCode(), test.NewUUID())
	assert.Nil(pts.T(), err)
}

func (pts *publisherIntegrationTestSuite) TestPublishFailureWhenEventCreationFails() {
	testCases := map[string]struct {
		eventCode string
		data      interface{}
	}{
		"test failure when event code is empty": {
			eventCode: test.EmptyString,
			data:      "some data",
		},
		"test failure when data is nil": {
			eventCode: pts.cfg.SignUpEventCode(),
			data:      nil,
		},
	}

	for name, testCase := range testCases {
		pts.Run(name, func() {
			err := pts.pb.Publish(testCase.eventCode, testCase.data)
			assert.Error(pts.T(), err)
		})
	}
}

func purgeMessages(t *testing.T) {
	cfg := config.NewConfig("../../local.env")

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

func TestPublisherIntegration(t *testing.T) {
	suite.Run(t, new(publisherIntegrationTestSuite))
}
