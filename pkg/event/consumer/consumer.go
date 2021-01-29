package consumer

import (
	"fmt"
	"github.com/streadway/amqp"
	"identification-service/pkg/config"
	"identification-service/pkg/event"
	"identification-service/pkg/liberr"
	"identification-service/pkg/queue"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"os"
	"os/signal"
	"syscall"
)

type Consumer interface {
	Start()
	consume(queueName string) error
}

type queueConsumer struct {
	lgr reporters.Logger
	cfg config.Config
	qu  queue.AMQP
	ss  session.Service
}

//TODO: REFACTOR THIS ENTIRE FILE
func (qc *queueConsumer) Start() {

	qc.lgr.InfoF("started consumer on %s", qc.cfg.AMPQConfig().Address())

	for _, queueName := range qc.cfg.ConsumerConfig().QueueNames() {
		go func(queueName string) {

			err := qc.consume(queueName)
			if err != nil {
				qc.lgr.Error(fmt.Sprintf("failed to consume message from %s: %v", queueName, err))
			}

		}(queueName)
	}

	handleGracefulShutdown(qc.lgr, qc.qu)
}

func (qc *queueConsumer) consume(queueName string) error {
	ch, err := qc.qu.Stream(queueName)
	if err != nil {
		return liberr.WithOp("Consumer.consume", err)
	}

	for {
		data := <-ch

		if err := handleData(qc.ss, data); err != nil {
			qc.lgr.Error(fmt.Sprintf("failed to handle message %v from %s: %v", data, queueName, err))
		}
	}
}

func handleData(ss session.Service, data amqp.Delivery) error {
	ev, err := event.FromBytes(data.Body)
	if err != nil {
		return err
	}

	switch ev.Code {
	case event.UpdatePassword:
		return handleUpdatePassword(ss, ev)
	default:
		return fmt.Errorf("invalid event code %s", ev.Code)
	}
}

func handleGracefulShutdown(lgr reporters.Logger, qu queue.AMQP) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	defer func() { _ = lgr.Flush() }()

	err := qu.Close()
	if err != nil {
		lgr.Error(err.Error())
		return
	}

	lgr.Info("worker shutdown successful")
}

func NewConsumer(cfg config.Config, lgr reporters.Logger, qu queue.AMQP, ss session.Service) Consumer {
	return &queueConsumer{
		cfg: cfg,
		lgr: lgr,
		qu:  qu,
		ss:  ss,
	}
}
