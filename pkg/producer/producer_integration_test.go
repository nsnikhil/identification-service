package producer_test

import (
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/producer"
	"identification-service/pkg/test"
	"testing"
)

type producerTestSuite struct {
	producer producer.Producer
	consumer sarama.Consumer
	suite.Suite
}

func (pts *producerTestSuite) SetupSuite() {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	cl, err := sarama.NewClient([]string{"localhost:9092"}, cfg)
	pts.Require().NoError(err)

	sp, err := sarama.NewSyncProducerFromClient(cl)
	pts.Require().NoError(err)

	cs, err := sarama.NewConsumerFromClient(cl)
	pts.Require().NoError(err)

	pts.producer = producer.NewProducer(sp)
	pts.consumer = cs
}

func TestProducer(t *testing.T) {
	suite.Run(t, new(producerTestSuite))
}

func (pts *producerTestSuite) TestProduceAndClose() {
	topic := "identification-service-producer-test-topics"
	message := test.RandBytes(8)

	p, o, err := pts.producer.Produce(topic, message)
	pts.Require().NoError(err)

	pc, err := pts.consumer.ConsumePartition(topic, p, o)
	pts.Require().NoError(err)

	msg := <-pc.Messages()
	pts.Assert().Equal(message, msg.Value)

	pts.Require().NoError(pts.producer.Close())
}
