package util_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/http/internal/util"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCopyWriterWrite(t *testing.T) {
	resp := httptest.NewRecorder()

	cr := util.NewCopyWriter(resp)

	b := []byte("data")

	_, err := cr.Write(b)
	require.NoError(t, err)

	cb, err := cr.Body()
	require.NoError(t, err)

	assert.Equal(t, b, cb)
}

func TestCopyWriterHeader(t *testing.T) {
	resp := httptest.NewRecorder()

	cr := util.NewCopyWriter(resp)

	cr.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, cr.Code())
}
