package consumer_test

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/consumer"
	"identification-service/pkg/queue"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
	"time"
)

type consumerTestSuite struct {
	consumer consumer.Consumer
	queue    queue.Queue
	cfg      config.QueueConfig
	svc      *session.MockService
	suite.Suite
}

func (cts *consumerTestSuite) SetupSuite() {
	cfg := config.NewConfig("../../local.env")

	lgr := reporters.NewLogger(cfg.Env(), cfg.LogConfig().Level())

	ch, err := queue.NewHandler(cfg.QueueConfig()).GetChannel()
	cts.Require().NoError(err)

	qu := queue.NewQueue(ch)

	time.Sleep(time.Second)

	mockSessionService := &session.MockService{}
	mockSessionService.On(
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.Anything,
	).Return(nil)

	rt := consumer.NewMessageRouter(cfg.QueueConfig(), mockSessionService)

	cts.consumer = consumer.NewConsumer(cfg.QueueConfig(), lgr, qu, rt)
	cts.queue = qu
	cts.cfg = cfg.QueueConfig()
	cts.svc = mockSessionService

	go cts.consumer.Start()
	time.Sleep(time.Second)
}

func TestConsumer(t *testing.T) {
	suite.Run(t, new(consumerTestSuite))
}

func (cts *consumerTestSuite) TestConsumeUpdatePasswordQueue() {
	err := cts.queue.Push(cts.cfg.UpdatePasswordQueueName(), test.RandBytes(8))
	cts.Require().NoError(err)

	time.Sleep(time.Second)

	cts.svc.AssertCalled(cts.T(),
		"RevokeAllSessions",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.Anything,
	)

	cts.Assert().NoError(cts.consumer.Close())
}
