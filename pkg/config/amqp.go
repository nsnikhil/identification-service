package config

import "fmt"

type AMPQConfig struct {
	host     string
	port     string
	user     string
	password string
	vhost    string

	queueName string
}

func newAMPQConfig() AMPQConfig {
	return AMPQConfig{
		host:      getString("AMPQ_HOST"),
		port:      getString("AMPQ_PORT"),
		user:      getString("AMPQ_USER"),
		password:  getString("AMPQ_PASSWORD"),
		vhost:     getString("AMPQ_VHOST"),
		queueName: getString("AMPQ_QUEUE_NAME"),
	}
}

func (ac AMPQConfig) QueueName() string {
	return ac.queueName
}

func (ac AMPQConfig) Address() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", ac.user, ac.password, ac.host, ac.port, ac.vhost)
}
