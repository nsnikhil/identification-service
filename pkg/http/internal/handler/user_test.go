package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	"identification-service/pkg/liberr"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/user"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	userName        = "Nikhil Soni"
	userEmail       = "n.nikhil.ns65@gmail.com"
	userPassword    = "Password@1234"
	userNewPassword = "NewPassword@1234"
	userID          = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
)

func TestCreateUserSuccess(t *testing.T) {
	service := &user.MockService{}
	service.On("CreateUser", userName, userEmail, userPassword).Return(userID, nil)

	req := contract.CreateUserRequest{Name: userName, Email: userEmail, Password: userPassword}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	expectedBody := `{"data":{"message":"user created successfully"},"success":true}`

	testCreateUser(t, http.StatusCreated, expectedBody, bytes.NewBuffer(b), service)
}

func TestCreateUserFailure(t *testing.T) {
	testCases := map[string]struct {
		service      func() user.Service
		body         func() io.Reader
		expectedCode int
		expectedBody string
	}{
		"test failure when body parsing fails": {
			service:      func() user.Service { return &user.MockService{} },
			body:         func() io.Reader { return nil },
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":{"message":"unexpected end of JSON input"},"success":false}`,
		},
		"test failure when service call fails fails": {
			service: func() user.Service {
				service := &user.MockService{}
				service.On("CreateUser", userName, userEmail, userPassword).Return("", liberr.WithArgs(errors.New("failed to create new user")))

				return service
			},
			body: func() io.Reader {
				req := contract.CreateUserRequest{Name: userName, Email: userEmail, Password: userPassword}

				b, err := json.Marshal(req)
				require.NoError(t, err)

				return bytes.NewBuffer(b)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"error":{"message":"internal server error"},"success":false}`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCreateUser(t, testCase.expectedCode, testCase.expectedBody, testCase.body(), testCase.service())
		})
	}
}

func testCreateUser(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service user.Service) {
	lgr := reporters.NewLogger("dev", "debug")

	uh := handler.NewUserHandler(service)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/user/create", body)

	mdl.WithError(lgr, uh.SignUp)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}

func TestUpdatePasswordSuccess(t *testing.T) {
	mockUserService := &user.MockService{}
	mockUserService.On("UpdatePassword", userName, userPassword, userNewPassword).Return(nil)

	req := contract.UpdatePasswordRequest{Email: userName, OldPassword: userPassword, NewPassword: userNewPassword}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	expectedBody := `{"data":{"message":"password updated successfully"},"success":true}`

	testUpdatePassword(t, http.StatusOK, expectedBody, bytes.NewBuffer(b), mockUserService)
}

func TestUpdatePasswordFailureWhenSvcCallFails(t *testing.T) {
	mockUserService := &user.MockService{}
	mockUserService.On("UpdatePassword", userName, userPassword, userNewPassword).Return(liberr.WithArgs(errors.New("failed to update password")))

	req := contract.UpdatePasswordRequest{Email: userName, OldPassword: userPassword, NewPassword: userNewPassword}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testUpdatePassword(t, http.StatusInternalServerError, expectedBody, bytes.NewBuffer(b), mockUserService)
}

func testUpdatePassword(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service user.Service) {
	lgr := reporters.NewLogger("dev", "debug")

	uh := handler.NewUserHandler(service)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/user/update-password", body)

	mdl.WithError(lgr, uh.UpdatePassword)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}
