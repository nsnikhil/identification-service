package publisher_test

import (
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/event"
	"identification-service/pkg/event/publisher"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"testing"
	"time"
)

type publisherIntegrationTestSuite struct {
	qu queue.AMQP
	pb publisher.Publisher
	suite.Suite
}

func (pts *publisherIntegrationTestSuite) SetupSuite() {
	cfg := config.NewConfig("../../../local.env")

	qu := queue.NewAMQP(cfg.AMPQConfig().Address())
	time.Sleep(time.Millisecond * 100)

	pb, err := publisher.NewPublisher(qu, cfg.PublisherConfig().QueueMap())
	require.NoError(pts.T(), err)

	pts.qu = qu
	pts.pb = pb
}

func (pts *publisherIntegrationTestSuite) TearDownSuite() {
	require.NoError(pts.T(), pts.qu.Close())
}

func (pts *publisherIntegrationTestSuite) AfterTest(suiteName, testName string) {
	purgeMessages(pts.T())
}

func (pts *publisherIntegrationTestSuite) TestPublisherPublishSuccess() {
	err := pts.pb.Publish(event.SignUp, test.NewUUID())
	assert.Nil(pts.T(), err)
}

func (pts *publisherIntegrationTestSuite) TestPublishFailureWhenEventCreationFails() {
	testCases := map[string]struct {
		eventCode event.Code
		data      interface{}
	}{
		"test failure when event code is invalid": {
			eventCode: event.Code("other"),
			data:      "some data",
		},
		"test failure when data is nil": {
			eventCode: event.SignUp,
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
	cfg := config.NewConfig("../../../local.env")

	conn, err := amqp.Dial(cfg.AMPQConfig().Address())
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 100)

	ch, err := conn.Channel()
	require.NoError(t, err)

	for _, queueName := range cfg.PublisherConfig().QueueMap() {
		_, err = ch.QueuePurge(queueName, true)
		require.NoError(t, err)
	}
}

func TestPublisherIntegration(t *testing.T) {
	suite.Run(t, new(publisherIntegrationTestSuite))
}
