package producer

import (
	"github.com/Shopify/sarama"
	"github.com/nsnikhil/erx"
	"math"
)

type Producer interface {
	Produce(topic string, value []byte) (int32, int64, error)
	Close() error
}

type kafkaProducer struct {
	producer sarama.SyncProducer
}

func (kp *kafkaProducer) Produce(topic string, value []byte) (int32, int64, error) {
	partition, offset, err := kp.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	})

	if err != nil {
		return math.MinInt32, math.MinInt64, erx.WithArgs(erx.Operation("Producer.Produce"), erx.ProducerError, err)
	}

	return partition, offset, nil
}

func (kp *kafkaProducer) Close() error {
	if err := kp.producer.Close(); err != nil {
		return erx.WithArgs(erx.Operation("Producer.Close"), erx.ProducerError, err)
	}

	return nil
}

func NewProducer(producer sarama.SyncProducer) Producer {
	return &kafkaProducer{
		producer: producer,
	}
}
