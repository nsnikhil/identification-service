package config

import "github.com/stretchr/testify/mock"

type KafkaConfig interface {
	SignUpTopicName() string
	UpdatePasswordTopicName() string
	Addresses() []string
}

type appKafkaConfig struct {
	signUpTopicName         string
	updatePasswordTopicName string
	addresses               []string
}

func (akc appKafkaConfig) SignUpTopicName() string {
	return akc.signUpTopicName
}
func (akc appKafkaConfig) UpdatePasswordTopicName() string {
	return akc.updatePasswordTopicName
}

func (akc appKafkaConfig) Addresses() []string {
	return akc.addresses
}

func newKafkaConfig() KafkaConfig {
	return appKafkaConfig{
		addresses:               getStringSlice("KAFKA_ADDRESSES"),
		signUpTopicName:         getString("SIGNUP_EVENT_TOPIC_NAME"),
		updatePasswordTopicName: getString("UPDATE_PASSWORD_EVENT_TOPIC_NAME"),
	}
}

type MockKafkaConfig struct {
	mock.Mock
}

func (mock *MockKafkaConfig) SignUpTopicName() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockKafkaConfig) UpdatePasswordTopicName() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockKafkaConfig) Addresses() []string {
	args := mock.Called()
	return args.Get(0).([]string)
}
