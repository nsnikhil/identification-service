package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	userEmail := test.NewEmail()
	accessToken := test.NewPasetoToken()
	refreshToken := test.NewUUID()
	userPassword := test.NewPassword()

	reqBody := contract.LoginRequest{Email: userEmail, Password: userPassword}

	expectedBody := fmt.Sprintf(
		`{"data":{"access_token":"%s","refresh_token":"%s"},"success":true}`,
		accessToken,
		refreshToken,
	)

	mockSessionService := &session.MockService{}
	mockSessionService.On("LoginUser", mock.AnythingOfType("*context.emptyCtx"), userEmail, userPassword).Return(accessToken, refreshToken, nil)

	testLogin(t, http.StatusCreated, expectedBody, mockSessionService, reqBody)
}

func TestLoginFailure(t *testing.T) {
	userEmail := test.NewEmail()
	userPassword := test.NewPassword()

	reqBody := contract.LoginRequest{Email: userEmail, Password: userPassword}

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	mockSessionService := &session.MockService{}
	mockSessionService.On("LoginUser", mock.AnythingOfType("*context.emptyCtx"), userEmail, userPassword).Return("", "", liberr.WithArgs(errors.New("failed to login")))

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
	accessToken := test.NewPasetoToken()
	refreshToken := test.NewUUID()

	reqBody := contract.RefreshTokenRequest{RefreshToken: refreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("RefreshToken", mock.AnythingOfType("*context.emptyCtx"), refreshToken).Return(accessToken, nil)

	expectedBody := fmt.Sprintf(`{"data":{"access_token":"%s"},"success":true}`, accessToken)

	testRefreshToken(t, http.StatusOK, expectedBody, mockSessionService, reqBody)
}

func TestRefreshTokenFailure(t *testing.T) {
	refreshToken := test.NewUUID()

	reqBody := contract.RefreshTokenRequest{RefreshToken: refreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("RefreshToken", mock.AnythingOfType("*context.emptyCtx"), refreshToken).Return("", liberr.WithArgs(errors.New("failed to refresh token")))

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
	refreshToken := test.NewUUID()

	reqBody := contract.LogoutRequest{RefreshToken: refreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("LogoutUser", mock.AnythingOfType("*context.emptyCtx"), refreshToken).Return(nil)

	expectedBody := `{"data":{"message":"Logout Successful"},"success":true}`

	testLogout(t, http.StatusOK, expectedBody, mockSessionService, reqBody)
}

func TestLogoutFailure(t *testing.T) {
	refreshToken := test.NewUUID()

	reqBody := contract.LogoutRequest{RefreshToken: refreshToken}

	mockSessionService := &session.MockService{}
	mockSessionService.On("LogoutUser", mock.AnythingOfType("*context.emptyCtx"), refreshToken).Return(liberr.WithArgs(errors.New("failed to logout user")))

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
