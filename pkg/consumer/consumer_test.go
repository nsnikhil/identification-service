package consumer_test

import (
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/consumer"
	"identification-service/pkg/queue"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
)

//TODO: COMPLETE THIS
type consumerSuite struct {
	c consumer.Consumer
	suite.Suite
}

func (cs *consumerSuite) SetupSuite() {
	cfg := &config.MockConfig{}
	lgr := &reporters.MockLogger{}
	qu := &queue.MockAMQP{}
	ss := &session.MockService{}

	cs.c = consumer.NewConsumer(cfg, lgr, qu, ss)
}

func (cs *consumerSuite) TearDownSuite() {

}

func (cs *consumerSuite) AfterTest(suiteName, testName string) {

}

func (cs *consumerSuite) TestConsumerHandle() {

}
