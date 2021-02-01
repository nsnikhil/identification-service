package config

import "github.com/stretchr/testify/mock"

type EventConfig interface {
	SignUpEventCode() string
	SignUpQueueName() string
	UpdatePasswordEventCode() string
	UpdatePasswordQueueName() string
	QueueMap() map[string]string
}
type appEventConfig struct {
	signUpEventCode         string
	signUpQueueName         string
	updatePasswordEventCode string
	updatePasswordQueueName string
	queueMap                map[string]string
}

func (aec appEventConfig) SignUpEventCode() string {
	return aec.signUpEventCode
}

func (aec appEventConfig) SignUpQueueName() string {
	return aec.signUpQueueName
}

func (aec appEventConfig) UpdatePasswordEventCode() string {
	return aec.updatePasswordEventCode
}

func (aec appEventConfig) UpdatePasswordQueueName() string {
	return aec.updatePasswordQueueName
}

func (aec appEventConfig) QueueMap() map[string]string {
	return aec.queueMap
}

func newEventConfig() EventConfig {
	signUpEventCode := getString("SIGNUP_EVENT_CODE")
	signUpQueueName := getString("SIGNUP_EVENT_QUEUE_NAME")
	updatePasswordEventCode := getString("UPDATE_PASSWORD_EVENT_CODE")
	updatePasswordQueueName := getString("UPDATE_PASSWORD_EVENT_QUEUE_NAME")
	queueMap := map[string]string{
		signUpEventCode:         signUpQueueName,
		updatePasswordEventCode: updatePasswordQueueName,
	}

	return &appEventConfig{
		signUpEventCode:         signUpEventCode,
		signUpQueueName:         signUpQueueName,
		updatePasswordEventCode: updatePasswordEventCode,
		updatePasswordQueueName: updatePasswordQueueName,
		queueMap:                queueMap,
	}
}

type MockEventConfig struct {
	mock.Mock
}

func (mock *MockEventConfig) SignUpEventCode() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockEventConfig) SignUpQueueName() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockEventConfig) UpdatePasswordEventCode() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockEventConfig) UpdatePasswordQueueName() string {
	args := mock.Called()
	return args.String(0)
}

func (mock *MockEventConfig) QueueMap() map[string]string {
	args := mock.Called()
	return args.Get(0).(map[string]string)
}
