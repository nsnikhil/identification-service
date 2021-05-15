package consumer

import (
	"github.com/nsnikhil/erx"
	"identification-service/pkg/config"
	"identification-service/pkg/queue"
	reporters "identification-service/pkg/reporting"
	"os"
	"os/signal"
	"syscall"
)

type Consumer interface {
	Start()
	Close() error
}

type ampqConsumer struct {
	lgr           reporters.Logger
	cfg           config.QueueConfig
	queue         queue.Queue
	messageRouter MessageRouter
}

func (aq *ampqConsumer) Start() {
	go consume(aq.cfg.UpdatePasswordQueueName(), aq)
	handleGracefulShutdown(aq)
}

func handleGracefulShutdown(aq *ampqConsumer) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	defer func() { _ = aq.lgr.Flush() }()

	if err := aq.queue.Close(); err != nil {
		logError(
			erx.WithArgs(erx.Operation("Consumer.handleGracefulShutdown"), erx.ConsumerError, err),
			aq.lgr,
		)
	}

	aq.lgr.Info("consumer shutdown successful")
}

func (aq *ampqConsumer) Close() error {
	if err := aq.queue.Close(); err != nil {
		return erx.WithArgs(erx.Operation("Consumer.Close"), erx.ConsumerError, err)
	}

	return nil
}

func consume(queueName string, aq *ampqConsumer) {
	wrap := func(err error) error {
		return erx.WithArgs(erx.Operation("consume"), erx.ConsumerError, err)
	}

	dc, err := aq.queue.Consume(queueName)
	if err != nil {
		logError(wrap(err), aq.lgr)
	}

	for {
		msg := <-dc
		err = aq.messageRouter.Route(queueName, msg.Body)
		if err != nil {
			logError(wrap(err), aq.lgr)
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

func NewConsumer(
	cfg config.QueueConfig,
	lgr reporters.Logger,
	queue queue.Queue,
	messageRouter MessageRouter,
) Consumer {
	return &ampqConsumer{
		cfg:           cfg,
		lgr:           lgr,
		queue:         queue,
		messageRouter: messageRouter,
	}
}
