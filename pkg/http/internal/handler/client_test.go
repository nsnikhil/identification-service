package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	"identification-service/pkg/liberr"
	reporters "identification-service/pkg/reporting"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	clientName     = "clientOne"
	accessTokenTTL = 10
	sessionTTL     = 14440
	clientSecret   = "ce36dc88-0a27-498a-aef4-5051a7fd6e7f"
	clientID       = "f14abb31-ec1a-4ff6-a937-c2e930ca34ef"
)

func TestClientHandlerCreateSuccess(t *testing.T) {
	req := contract.CreateClientRequest{Name: clientName, AccessTokenTTL: accessTokenTTL, SessionTTL: sessionTTL}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On("CreateClient", mock.AnythingOfType("*context.emptyCtx"), clientName, accessTokenTTL, sessionTTL).Return(clientSecret, nil)

	expectedBody := `{"data":{"secret":"ce36dc88-0a27-498a-aef4-5051a7fd6e7f"},"success":true}`

	testClientHandlerCreate(t, http.StatusCreated, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func TestClientHandlerCreateFailure(t *testing.T) {
	req := contract.CreateClientRequest{Name: clientName, AccessTokenTTL: accessTokenTTL, SessionTTL: sessionTTL}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On("CreateClient", mock.AnythingOfType("*context.emptyCtx"), clientName, accessTokenTTL, sessionTTL).Return("", liberr.WithArgs(errors.New("failed to create client")))

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testClientHandlerCreate(t, http.StatusInternalServerError, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func testClientHandlerCreate(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service client.Service) {
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/client/create", body)

	ch := handler.NewClientHandler(service)

	mdl.WithError(reporters.NewLogger("dev", "debug"), ch.Register)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}

func TestClientRevokeSuccess(t *testing.T) {
	req := contract.ClientRevokeRequest{ID: clientID}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On("RevokeClient", mock.AnythingOfType("*context.emptyCtx"), clientID).Return(nil)

	expectedBody := `{"data":{"message":"client revoked successfully"},"success":true}`

	testClientHandlerRevoke(t, http.StatusOK, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func TestClientRevokeFailure(t *testing.T) {
	req := contract.ClientRevokeRequest{ID: clientID}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On("RevokeClient", mock.AnythingOfType("*context.emptyCtx"), clientID).Return(liberr.WithArgs(errors.New("failed to revoke client")))

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testClientHandlerRevoke(t, http.StatusInternalServerError, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func testClientHandlerRevoke(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service client.Service) {
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/client/revoke", body)

	ch := handler.NewClientHandler(service)

	mdl.WithError(reporters.NewLogger("dev", "debug"), ch.Revoke)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}
