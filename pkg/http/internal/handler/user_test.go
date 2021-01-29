package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	"identification-service/pkg/liberr"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateUserSuccess(t *testing.T) {
	userName, userEmail, userPassword := test.RandString(8), test.NewEmail(), test.NewPassword()

	service := &user.MockService{}
	service.On("CreateUser", mock.AnythingOfType("*context.emptyCtx"), userName, userEmail, userPassword).Return(test.NewUUID(), nil)

	req := contract.CreateUserRequest{Name: userName, Email: userEmail, Password: userPassword}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	expectedBody := `{"data":{"message":"user created successfully"},"success":true}`

	testCreateUser(t, http.StatusCreated, expectedBody, bytes.NewBuffer(b), service)
}

func TestCreateUserFailure(t *testing.T) {
	userName, userEmail, userPassword := test.RandString(8), test.NewEmail(), test.NewPassword()

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
				service.On("CreateUser", mock.AnythingOfType("*context.emptyCtx"), userName, userEmail, userPassword).Return("", liberr.WithArgs(errors.New("failed to create new user")))

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

	mdl.WithErrorHandler(lgr, uh.SignUp)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}

func TestUpdatePasswordSuccess(t *testing.T) {
	userEmail, userPassword, userPasswordNew := test.NewEmail(), test.NewPassword(), test.NewPassword()

	mockUserService := &user.MockService{}
	mockUserService.On("UpdatePassword", mock.AnythingOfType("*context.emptyCtx"), userEmail, userPassword, userPasswordNew).Return(nil)

	req := contract.UpdatePasswordRequest{Email: userEmail, OldPassword: userPassword, NewPassword: userPasswordNew}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	expectedBody := `{"data":{"message":"password updated successfully"},"success":true}`

	testUpdatePassword(t, http.StatusOK, expectedBody, bytes.NewBuffer(b), mockUserService)
}

func TestUpdatePasswordFailureWhenSvcCallFails(t *testing.T) {
	userEmail, userPassword, userPasswordNew := test.NewEmail(), test.NewPassword(), test.NewPassword()

	mockUserService := &user.MockService{}
	mockUserService.On("UpdatePassword", mock.AnythingOfType("*context.emptyCtx"), userEmail, userPassword, userPasswordNew).Return(liberr.WithArgs(errors.New("failed to update password")))

	req := contract.UpdatePasswordRequest{Email: userEmail, OldPassword: userPassword, NewPassword: userPasswordNew}

	b, err := json.Marshal(req)
	require.NoError(t, err)

	expectedBody := `{"error":{"message":"internal server error"},"success":false}`

	testUpdatePassword(t, http.StatusInternalServerError, expectedBody, bytes.NewBuffer(b), mockUserService)
}

func TestUpdatePasswordFailureWhenValidationFails(t *testing.T) {
	userEmail, userPassword, userPasswordNew := test.NewEmail(), test.NewPassword(), test.NewPassword()

	expectedBody := func(message string) string {
		return fmt.Sprintf(`{"error":{"message":"%s"},"success":false}`, message)
	}

	toReader := func(reqBody contract.UpdatePasswordRequest) io.Reader {
		b, err := json.Marshal(reqBody)
		require.NoError(t, err)

		return bytes.NewBuffer(b)
	}

	testCases := map[string]struct {
		reqBody contract.UpdatePasswordRequest
		errMsg  string
	}{
		"test failure when email is empty": {
			reqBody: contract.UpdatePasswordRequest{
				Email:       test.EmptyString,
				OldPassword: userPassword,
				NewPassword: userPasswordNew,
			},
			errMsg: "email cannot be empty",
		},
		"test failure when old password is empty": {
			reqBody: contract.UpdatePasswordRequest{
				Email:       userEmail,
				OldPassword: test.EmptyString,
				NewPassword: userPasswordNew,
			},
			errMsg: "old password cannot be empty",
		},
		"test failure when new password is empty": {
			reqBody: contract.UpdatePasswordRequest{
				Email:       userEmail,
				OldPassword: userPassword,
				NewPassword: test.EmptyString,
			},
			errMsg: "new password cannot be empty",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testUpdatePassword(t, http.StatusBadRequest, expectedBody(testCase.errMsg), toReader(testCase.reqBody), &user.MockService{})
		})
	}
}

func testUpdatePassword(t *testing.T, expectedCode int, expectedBody string, body io.Reader, service user.Service) {
	lgr := reporters.NewLogger("dev", "debug")

	uh := handler.NewUserHandler(service)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/user/update-password", body)

	mdl.WithErrorHandler(lgr, uh.UpdatePassword)(w, r)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedBody, w.Body.String())
}
