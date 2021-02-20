package consumer_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/config"
	"identification-service/pkg/consumer"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
)

func TestRouterSuccess(t *testing.T) {
	userID := test.NewUUID()

	mockKafkaConfig := &config.MockKafkaConfig{}
	mockKafkaConfig.On("UpdatePasswordTopicName").Return("update-password")

	mockSessionService := &session.MockService{}
	mockSessionService.On("RevokeAllSessions", mock.AnythingOfType("*context.emptyCtx"), userID).
		Return(nil)

	rt := consumer.NewMessageRouter(mockKafkaConfig, mockSessionService)

	testCases := map[string]struct {
		topic string
		msg   []byte
	}{
		"test router update password topic": {
			topic: "update-password",
			msg:   []byte(userID),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, rt.Route(testCase.topic, testCase.msg))
		})
	}
}

func TestRouterFailure(t *testing.T) {
	userID := test.NewUUID()

	testCases := map[string]struct {
		topic string
		msg   []byte
		cfg   func() config.KafkaConfig
		ss    func() session.Service
	}{
		"test router failure when topic name is invalid": {
			topic: test.RandString(89),
			msg:   []byte(userID),
			cfg: func() config.KafkaConfig {
				mockKafkaConfig := &config.MockKafkaConfig{}
				mockKafkaConfig.On("UpdatePasswordTopicName").Return("update-password")
				return mockKafkaConfig
			},
			ss: func() session.Service { return &session.MockService{} },
		},
		"test router failure when handler returns error": {
			topic: "update-password",
			msg:   []byte(userID),
			cfg: func() config.KafkaConfig {
				mockKafkaConfig := &config.MockKafkaConfig{}
				mockKafkaConfig.On("UpdatePasswordTopicName").Return("update-password")
				return mockKafkaConfig
			},
			ss: func() session.Service {
				mockSessionService := &session.MockService{}
				mockSessionService.On("RevokeAllSessions", mock.AnythingOfType("*context.emptyCtx"), userID).
					Return(errors.New("failed to revoke all sessions"))
				return mockSessionService
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			rt := consumer.NewMessageRouter(testCase.cfg(), testCase.ss())
			assert.Error(t, rt.Route(testCase.topic, testCase.msg))
		})
	}
}
