package config

type ConsumerConfig struct {
	queueNames []string
}

func newConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		queueNames: getStringSlice("CONSUMER_QUEUE_NAMES"),
	}
}

func (cc ConsumerConfig) QueueNames() []string {
	return cc.queueNames
}
