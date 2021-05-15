package config

import (
	"fmt"
	"github.com/stretchr/testify/mock"
)

type QueueConfig interface {
	SignUpQueueName() string
	UpdatePasswordQueueName() string
	Address() string
}

type appQueueConfig struct {
	host                    string
	port                    string
	user                    string
	password                string
	vhost                   string
	signUpQueueName         string
	updatePasswordQueueName string
}

func newQueueConfig() QueueConfig {
	return appQueueConfig{
		host:                    getString("AMPQ_HOST"),
		port:                    getString("AMPQ_PORT"),
		user:                    getString("AMPQ_USER"),
		password:                getString("AMPQ_PASSWORD"),
		vhost:                   getString("AMPQ_VHOST"),
		signUpQueueName:         getString("SIGNUP_EVENT_QUEUE_NAME"),
		updatePasswordQueueName: getString("UPDATE_PASSWORD_EVENT_QUEUE_NAME"),
	}
}

func (qc appQueueConfig) SignUpQueueName() string {
	return qc.signUpQueueName
}

func (qc appQueueConfig) UpdatePasswordQueueName() string {
	return qc.updatePasswordQueueName
}

func (qc appQueueConfig) Address() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s", qc.user, qc.password, qc.host, qc.port, qc.vhost)
}

type MockQueueConfig struct {
	mock.Mock
}

func (mock *MockQueueConfig) SignUpQueueName() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockQueueConfig) UpdatePasswordQueueName() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockQueueConfig) Address() string {
	args := mock.Called()
	return args.String(0)
}
