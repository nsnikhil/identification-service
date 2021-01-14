package test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/http/contract"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	deps := setupTest(t)

	req := newRequest(t, http.MethodGet, "ping", nil)
	resp := execRequest(t, deps.cl, req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var data contract.APIResponse

	err = json.Unmarshal(b, &data)
	require.NoError(t, err)

	assert.True(t, data.Success)
	assert.Equal(t, "pong", data.Data)
}
