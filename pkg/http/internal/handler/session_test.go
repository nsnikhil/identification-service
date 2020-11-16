package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	"identification-service/pkg/liberr"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginSuccess(t *testing.T) {
	reqBody := contract.LoginRequest{Email: test.UserEmail, Password: test.UserPassword}

	expectedBody := `{"data":{"access_token":"v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMDozNjowNyswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTA6MjY6MDcrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiMTEwMTI0NjUtMDNhNC00OWI2LTgwODEtY2RmYzczMDlhY2MwIiwibmJmIjoiMjAyMC0xMS0wN1QxMDoyNjowNyswNTozMCJ9PrXViH5779NxXHK_PxnwW-FdFV0klU07umd8X7F0A9irFLX7GTS3AczNm_hmb_yfYOX0o4DJri89AWeCb0qTAg.bnVsbA","refresh_token":"5df8159e-fd51-4e6c-9849-a9b1f070a403"},"success":true}`

	mockSessionService := &session.MockService{}
	mockSessionService.On("LoginUser", mock.AnythingOfType("*context.emptyCtx"), test.UserEmail, test.UserPassword).Return(test.SessionAccessToken, test.SessionRefreshToken, nil)

	testLogin(t, http.StatusCreated, expectedBody, mockSessionService, reqBody)
}

func TestLoginFailure(t *testing.T) {
	reqBody := contract.LoginRequest{Email: test.UserEmail, Password: test.UserPassword}

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	mockSessionService := &session.MockService{}
	mockSessionService.On("LoginUser", mock.AnythingOfType("*context.emptyCtx"), test.UserEmail, test.UserPassword).Return("", "", liberr.WithArgs(errors.New("failed to login")))

	testLogin(t, http.StatusInternalServerError, expectedBody, mockSessionService, reqBody)
}

func testLogin(t *testing.T, expectedCode int, expectedBody string, sessionService session.Service, reqBody contract.LoginRequest) {
	b, err := json.Marshal(&reqBody)

	r, err := http.NewRequest(http.MethodPost, "/session/login", bytes.NewBuffer(b))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	sh := handler.NewSessionHandler(sessionService)

	lgr := reporters.NewLogger("dev", "debug")
	mdl.WithErrorHandler(lgr, sh.Login)(w, r)

	require.Equal(t, expectedCode, w.Code)

	assert.Equal(t, expectedBody, w.Body.String())
}

func TestRefreshTokenSuccess(t *testing.T) {
	reqBody := contract.RefreshTokenRequest{RefreshToken: test.SessionRefreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("RefreshToken", mock.AnythingOfType("*context.emptyCtx"), test.SessionRefreshToken).Return(test.SessionAccessToken, nil)

	expectedBody := `{"data":{"access_token":"v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMDozNjowNyswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTA6MjY6MDcrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiMTEwMTI0NjUtMDNhNC00OWI2LTgwODEtY2RmYzczMDlhY2MwIiwibmJmIjoiMjAyMC0xMS0wN1QxMDoyNjowNyswNTozMCJ9PrXViH5779NxXHK_PxnwW-FdFV0klU07umd8X7F0A9irFLX7GTS3AczNm_hmb_yfYOX0o4DJri89AWeCb0qTAg.bnVsbA"},"success":true}`

	testRefreshToken(t, http.StatusOK, expectedBody, mockSessionService, reqBody)
}

func TestRefreshTokenFailure(t *testing.T) {
	reqBody := contract.RefreshTokenRequest{RefreshToken: test.SessionRefreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("RefreshToken", mock.AnythingOfType("*context.emptyCtx"), test.SessionRefreshToken).Return("", liberr.WithArgs(errors.New("failed to refresh token")))

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testRefreshToken(t, http.StatusInternalServerError, expectedBody, mockSessionService, reqBody)
}

func testRefreshToken(t *testing.T, expectedCode int, expectedBody string, sessionService session.Service, reqBody contract.RefreshTokenRequest) {
	b, err := json.Marshal(&reqBody)

	r, err := http.NewRequest(http.MethodPost, "/session/refresh-token", bytes.NewBuffer(b))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	sh := handler.NewSessionHandler(sessionService)

	lgr := reporters.NewLogger("dev", "debug")
	mdl.WithErrorHandler(lgr, sh.RefreshToken)(w, r)

	require.Equal(t, expectedCode, w.Code)
	require.Equal(t, expectedBody, w.Body.String())
}

func TestLogoutSuccess(t *testing.T) {
	reqBody := contract.LogoutRequest{RefreshToken: test.SessionRefreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("LogoutUser", mock.AnythingOfType("*context.emptyCtx"), test.SessionRefreshToken).Return(nil)

	expectedBody := `{"data":{"message":"Logout Successful"},"success":true}`

	testLogout(t, http.StatusOK, expectedBody, mockSessionService, reqBody)
}

func TestLogoutFailure(t *testing.T) {
	reqBody := contract.LogoutRequest{RefreshToken: test.SessionRefreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("LogoutUser", mock.AnythingOfType("*context.emptyCtx"), test.SessionRefreshToken).Return(liberr.WithArgs(errors.New("failed to logout user")))

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testLogout(t, http.StatusInternalServerError, expectedBody, mockSessionService, reqBody)
}

func testLogout(t *testing.T, expectedCode int, expectedBody string, sessionService session.Service, reqBody contract.LogoutRequest) {
	b, err := json.Marshal(&reqBody)

	r, err := http.NewRequest(http.MethodPost, "/session/refresh-token", bytes.NewBuffer(b))
	require.NoError(t, err)

	w := httptest.NewRecorder()

	sh := handler.NewSessionHandler(sessionService)

	lgr := reporters.NewLogger("dev", "debug")
	mdl.WithErrorHandler(lgr, sh.Logout)(w, r)

	require.Equal(t, expectedCode, w.Code)
	require.Equal(t, expectedBody, w.Body.String())
}
