package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nsnikhil/erx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/test"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHandlerCreateSuccess(t *testing.T) {
	clientName := test.RandString(8)
	clientEncodedPublicKey := test.RandString(44)
	clientSecret := test.NewUUID()
	accessTokenTTL := test.RandInt(1, 10)
	sessionTokenTTL := test.RandInt(1, 10)
	maxActiveSession := test.RandInt(1, 10)

	req := contract.CreateClientRequest{
		Name:              clientName,
		AccessTokenTTL:    accessTokenTTL,
		SessionTTL:        sessionTokenTTL,
		MaxActiveSessions: maxActiveSession,
		SessionStrategy:   test.ClientSessionStrategyRevokeOld,
	}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"CreateClient",
		mock.AnythingOfType("*context.emptyCtx"),
		clientName,
		accessTokenTTL,
		sessionTokenTTL,
		maxActiveSession,
		test.ClientSessionStrategyRevokeOld,
	).Return(clientEncodedPublicKey, clientSecret, nil)

	expectedBody := fmt.Sprintf(
		`{"data":{"public_key":"%s","secret":"%s"},"success":true}`,
		clientEncodedPublicKey,
		clientSecret,
	)

	testClientHandlerCreate(t, http.StatusCreated, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func TestClientHandlerCreateFailure(t *testing.T) {
	clientName := test.RandString(8)
	accessTokenTTL := test.RandInt(1, 10)
	sessionTokenTTL := test.RandInt(1, 10)
	maxActiveSession := test.RandInt(1, 10)

	req := contract.CreateClientRequest{
		Name:              clientName,
		AccessTokenTTL:    accessTokenTTL,
		SessionTTL:        sessionTokenTTL,
		MaxActiveSessions: maxActiveSession,
		SessionStrategy:   test.ClientSessionStrategyRevokeOld,
	}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"CreateClient",
		mock.AnythingOfType("*context.emptyCtx"),
		clientName,
		accessTokenTTL,
		sessionTokenTTL,
		maxActiveSession,
		test.ClientSessionStrategyRevokeOld,
	).Return("", "", erx.WithArgs(errors.New("failed to create client")))

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testClientHandlerCreate(t, http.StatusInternalServerError, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func testClientHandlerCreate(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service client.Service) {
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/client/create", body)

	ch := handler.NewClientHandler(service)

	mdl.WithErrorHandler(reporters.NewLogger("dev", "debug"), ch.Register)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}

func TestClientRevokeSuccess(t *testing.T) {
	clientID := test.NewUUID()

	req := contract.ClientRevokeRequest{ID: clientID}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"RevokeClient",
		mock.AnythingOfType("*context.emptyCtx"),
		clientID,
	).Return(nil)

	expectedBody := `{"data":{"message":"client revoked successfully"},"success":true}`

	testClientHandlerRevoke(t, http.StatusOK, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func TestClientRevokeFailure(t *testing.T) {
	clientID := test.NewUUID()

	req := contract.ClientRevokeRequest{ID: clientID}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"RevokeClient",
		mock.AnythingOfType("*context.emptyCtx"),
		clientID,
	).Return(erx.WithArgs(errors.New("failed to revoke client")))

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testClientHandlerRevoke(t, http.StatusInternalServerError, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func testClientHandlerRevoke(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service client.Service) {
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/client/revoke", body)

	ch := handler.NewClientHandler(service)

	mdl.WithErrorHandler(reporters.NewLogger("dev", "debug"), ch.Revoke)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}
