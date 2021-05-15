package queue_test

import (
	"github.com/stretchr/testify/require"
	"identification-service/pkg/config"
	"identification-service/pkg/queue"
	"testing"
)

func TestGetChannelSuccess(t *testing.T) {
	cfg := config.NewConfig("../../local.env").QueueConfig()

	_, err := queue.NewHandler(cfg).GetChannel()
	require.NoError(t, err)
}

func TestGetChannelFailure(t *testing.T) {
	cfg := &config.MockQueueConfig{}
	cfg.On("Address").Return("localhost:8080")

	_, err := queue.NewHandler(cfg).GetChannel()
	require.Error(t, err)
}
