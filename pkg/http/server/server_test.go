package server_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/config"
	"identification-service/pkg/http/server"
	reporters "identification-service/pkg/reporting"
	"net/http"
	"testing"
	"time"
)

func TestServerStart(t *testing.T) {
	cfg := config.NewConfig("../../../local.env")
	mockLogger := &reporters.MockLogger{}
	mockLogger.On("InfoF", []interface{}{"listening on ", ":8080"})

	rt := http.NewServeMux()
	rt.HandleFunc("/ping", func(resp http.ResponseWriter, req *http.Request) {})

	srv := server.NewServer(cfg, mockLogger, rt)
	go srv.Start()

	//TODO REMOVE SLEEP
	time.Sleep(time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:8080/ping")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
