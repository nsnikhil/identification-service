package app

import (
	"identification-service/pkg/config"
	"identification-service/pkg/consumer"
	"identification-service/pkg/queue"
	"time"
)

func StartWorker(configFile string) {
	cfg := config.NewConfig(configFile)

	lgr := initLogger(cfg)

	qu := queue.NewAMQP(cfg.AMPQConfig().Address())

	time.Sleep(time.Second)

	_, _, ss := initServices(cfg)

	consumer.NewConsumer(cfg, lgr, qu, ss).Start()
}
