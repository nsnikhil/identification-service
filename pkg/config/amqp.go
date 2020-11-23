package config

import "fmt"

type AMPQConfig struct {
	host     string
	port     string
	user     string
	password string
	vhost    string
}

func newAMPQConfig() AMPQConfig {
	return AMPQConfig{
		host:     getString("AMPQ_HOST"),
		port:     getString("AMPQ_PORT"),
		user:     getString("AMPQ_USER"),
		password: getString("AMPQ_PASSWORD"),
		vhost:    getString("AMPQ_VHOST"),
	}
}

func (ac AMPQConfig) Address() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", ac.user, ac.password, ac.host, ac.port, ac.vhost)
}
