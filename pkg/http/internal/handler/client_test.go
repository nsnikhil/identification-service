package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	"identification-service/pkg/liberr"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/test"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHandlerCreateSuccess(t *testing.T) {
	clientSecret := test.ClientSecret()

	req := contract.CreateClientRequest{
		Name:              test.ClientName(),
		AccessTokenTTL:    test.ClientAccessTokenTTL,
		SessionTTL:        test.ClientSessionTTL,
		MaxActiveSessions: test.ClientMaxActiveSessions,
		SessionStrategy:   test.ClientSessionStrategyRevokeOld,
	}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"CreateClient",
		mock.AnythingOfType("*context.emptyCtx"),
		test.ClientName(),
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
	).Return(test.ClientEncodedPublicKey, clientSecret, nil)

	expectedBody := fmt.Sprintf(
		`{"data":{"public_key":"8lchzCKRbdXEHsG/hJNMjMqdJLbIvAvDoViJtlcwWWo","secret":"%s"},"success":true}`,
		clientSecret,
	)

	testClientHandlerCreate(t, http.StatusCreated, expectedBody, bytes.NewBuffer(body), mockClientService)
}

func TestClientHandlerCreateFailure(t *testing.T) {
	req := contract.CreateClientRequest{
		Name:              test.ClientName(),
		AccessTokenTTL:    test.ClientAccessTokenTTL,
		SessionTTL:        test.ClientSessionTTL,
		MaxActiveSessions: test.ClientMaxActiveSessions,
		SessionStrategy:   test.ClientSessionStrategyRevokeOld,
	}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"CreateClient",
		mock.AnythingOfType("*context.emptyCtx"),
		test.ClientName(),
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
	).Return("", "", liberr.WithArgs(errors.New("failed to create client")))

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
	clientID := test.ClientID()

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
	clientID := test.ClientID()

	req := contract.ClientRevokeRequest{ID: clientID}

	body, err := json.Marshal(&req)
	require.NoError(t, err)

	mockClientService := &client.MockService{}
	mockClientService.On(
		"RevokeClient",
		mock.AnythingOfType("*context.emptyCtx"),
		clientID,
	).Return(liberr.WithArgs(errors.New("failed to revoke client")))

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
