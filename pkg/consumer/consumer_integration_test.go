package consumer_test

import (
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/consumer"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
	"time"
)

type consumerTestSuite struct {
	consumer consumer.Consumer
	producer sarama.SyncProducer
	cfg      config.KafkaConfig
	svc      *session.MockService
	suite.Suite
}

func (cts *consumerTestSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	lgr := reporters.NewLogger(cfg.Env(), cfg.LogConfig().Level())

	kCfg := sarama.NewConfig()
	kCfg.Producer.Return.Successes = true

	cl, err := sarama.NewClient(cfg.KafkaConfig().Addresses(), kCfg)
	cts.Require().NoError(err)

	cs, err := sarama.NewConsumerFromClient(cl)
	cts.Require().NoError(err)

	pd, err := sarama.NewSyncProducerFromClient(cl)
	cts.Require().NoError(err)

	mockSessionService := &session.MockService{}
	mockSessionService.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.Anything,
	).Return(nil)

	rt := consumer.NewMessageRouter(cfg.KafkaConfig(), mockSessionService)

	cts.consumer = consumer.NewConsumer(cfg.KafkaConfig(), lgr, cs, rt)
	cts.producer = pd
	cts.cfg = cfg.KafkaConfig()
	cts.svc = mockSessionService

	go cts.consumer.Start()
	time.Sleep(time.Second)
}

func TestConsumer(t *testing.T) {
	suite.Run(t, new(consumerTestSuite))
}

func (cts *consumerTestSuite) TestConsumeUpdatePasswordTopic() {
	msg := &sarama.ProducerMessage{
		Topic: cts.cfg.UpdatePasswordTopicName(),
		Value: sarama.ByteEncoder(test.NewUUID()),
	}

	_, _, err := cts.producer.SendMessage(msg)
	cts.Require().NoError(err)

	cts.svc.AssertCalled(cts.T(),
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.Anything,
	)

	cts.Assert().NoError(cts.consumer.Close())
}
