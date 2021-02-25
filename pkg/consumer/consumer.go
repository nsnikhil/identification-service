package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
	reporters "identification-service/pkg/reporting"
	"os"
	"os/signal"
	"syscall"
)

type Consumer interface {
	Start()
	Close() error
}

type kafkaConsumer struct {
	cfg           config.KafkaConfig
	lgr           reporters.Logger
	consumer      sarama.Consumer
	messageRouter MessageRouter
}

func (kc *kafkaConsumer) Start() {
	go consume(kc.cfg.UpdatePasswordTopicName(), kc)
	handleGracefulShutdown(kc)
}

func handleGracefulShutdown(kc *kafkaConsumer) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	defer func() { _ = kc.lgr.Flush() }()

	if err := kc.consumer.Close(); err != nil {
		logError(
			erx.WithArgs(erx.Operation("Consumer.handleGracefulShutdown"), erx.ConsumerError, err),
			kc.lgr,
		)
	}

	kc.lgr.Info("consumer shutdown successful")
}

func (kc *kafkaConsumer) Close() error {
	if err := kc.consumer.Close(); err != nil {
		return erx.WithArgs(erx.Operation("Consumer.Close"), erx.ConsumerError, err)
	}

	return nil
}

func consume(topic string, kc *kafkaConsumer) {
	wrap := func(err error) error {
		return erx.WithArgs(erx.Operation("consume"), erx.ConsumerError, err)
	}

	partitions, err := kc.consumer.Partitions(topic)
	if err != nil {
		logError(wrap(err), kc.lgr)
		return
	}

	pc, err := kc.consumer.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
	if err != nil {
		logError(wrap(err), kc.lgr)
		return
	}

	for {
		select {
		case msg := <-pc.Messages():
			if err := kc.messageRouter.Route(topic, msg.Value); err != nil {
				logError(wrap(err), kc.lgr)
			}
		case crr := <-pc.Errors():
			logError(wrap(crr.Err), kc.lgr)
		}
	}
}

func logError(err error, lgr reporters.Logger) {
	t, ok := err.(*erx.Erx)
	if ok {
		lgr.Error(t.String())
	} else {
		lgr.Error(err.Error())
	}
}

func NewConsumer(cfg config.KafkaConfig, lgr reporters.Logger, consumer sarama.Consumer, messageRouter MessageRouter) Consumer {
	return &kafkaConsumer{
		cfg:           cfg,
		lgr:           lgr,
		consumer:      consumer,
		messageRouter: messageRouter,
	}
}
