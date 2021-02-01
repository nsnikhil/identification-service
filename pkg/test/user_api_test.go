package test

import (
	"bytes"
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/http/contract"
	"net/http"
	"testing"
)

const (
	userNameKey        = "name"
	userEmailKey       = "email"
	userPasswordKey    = "password"
	userNewPasswordKey = "newPassword"
)

type userAPITestSuite struct {
	deps testDeps
	suite.Suite
}

func (uat *userAPITestSuite) SetupSuite() {
	uat.deps = setupTest(uat.T())
}

func (uat *userAPITestSuite) TearDownSuite() {
	truncateTables(uat.T(), uat.deps.ctx, uat.deps.db, ClientTableName, UserTableName)
	tearDownTest(uat.T(), uat.deps)
}

func TestUserAPI(t *testing.T) {
	suite.Run(t, new(userAPITestSuite))
}

func (uat *userAPITestSuite) TestSignUpUserSuccess() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})

	reqBody := getCreateUserReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "user created successfully"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.EventConfig(),
		uat.deps.cl,
		uat.deps.ch,
		http.StatusCreated,
		expectedRespData,
		authHeaders,
		reqBody,
		true,
	)
}

func (uat *userAPITestSuite) TestSignUpUserFailureWhenClientCredentialsAreMissing() {
	reqBody := getCreateUserReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "authentication failed"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.EventConfig(),
		uat.deps.cl,
		uat.deps.ch,
		http.StatusUnauthorized,
		expectedRespData,
		map[string]string{},
		reqBody,
		false,
	)
}

func (uat *userAPITestSuite) TestSignUpUserClientAuthenticationFailure() {
	defaultAuthHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})

	clientId := defaultAuthHeaders["CLIENT-ID"]
	clientSecret := defaultAuthHeaders["CLIENT-SECRET"]

	expectedRespData := func(msg string) contract.APIResponse {
		return contract.APIResponse{
			Success: false,
			Error:   &contract.Error{Message: msg},
		}
	}

	testCases := map[string]struct {
		authHeader map[string]string
	}{
		"test failure when client id is invalid": {
			authHeader: map[string]string{
				"CLIENT-ID":     "invalid",
				"CLIENT-SECRET": clientSecret,
			},
		},
		"test failure when client secret is invalid": {
			authHeader: map[string]string{
				"CLIENT-ID":     clientId,
				"CLIENT-SECRET": "invalid",
			},
		},
	}

	for name, testCase := range testCases {
		uat.T().Run(name, func(t *testing.T) {
			testSignUpUser(
				uat.T(),
				uat.deps.cfg.EventConfig(),
				uat.deps.cl,
				uat.deps.ch,
				http.StatusUnauthorized,
				expectedRespData("authentication failed"),
				testCase.authHeader,
				getCreateUserReqBody(map[string]interface{}{}),
				false,
			)
		})
	}
}

func (uat *userAPITestSuite) TestSignUpUserValidationFailure() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})

	expectedRespData := func(msg string) contract.APIResponse {
		return contract.APIResponse{
			Success: false,
			Error:   &contract.Error{Message: msg},
		}
	}

	testCases := map[string]struct {
		data         map[string]interface{}
		errorMessage string
	}{
		"test failure when name is empty": {
			data:         map[string]interface{}{userNameKey: EmptyString},
			errorMessage: "name cannot be empty",
		},
		"test failure when email is empty": {
			data:         map[string]interface{}{userEmailKey: EmptyString},
			errorMessage: "email cannot be empty",
		},
		"test failure when password is empty": {
			data:         map[string]interface{}{userPasswordKey: EmptyString},
			errorMessage: "password cannot be empty",
		},
		"test failure when password is below min characters": {
			data:         map[string]interface{}{userPasswordKey: RandString(6)},
			errorMessage: "password must be at least 8 characters long",
		},
		"test failure when password is invalid": {
			data:         map[string]interface{}{userPasswordKey: RandString(12)},
			errorMessage: "password must have at least 1 number, 1 lower character, 1 upper character and 1 symbol",
		},
	}

	for name, testCase := range testCases {
		uat.T().Run(name, func(t *testing.T) {
			testSignUpUser(
				uat.T(),
				uat.deps.cfg.EventConfig(),
				uat.deps.cl,
				uat.deps.ch,
				http.StatusBadRequest,
				expectedRespData(testCase.errorMessage),
				authHeaders,
				getCreateUserReqBody(testCase.data),
				false,
			)
		})
	}
}

func (uat *userAPITestSuite) TestSignUpUserFailureForDuplicateRecord() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})

	reqBody := getCreateUserReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "user created successfully"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.EventConfig(),
		uat.deps.cl,
		uat.deps.ch,
		http.StatusCreated,
		expectedRespData,
		authHeaders,
		reqBody,
		false,
	)

	expectedRespData = contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "duplicate record"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.EventConfig(),
		uat.deps.cl,
		uat.deps.ch,
		http.StatusConflict,
		expectedRespData,
		authHeaders,
		reqBody,
		false,
	)
}

func (uat *userAPITestSuite) TestUpdatePasswordSuccess() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(uat.T(), uat.deps.cfg.EventConfig(), uat.deps.cl, uat.deps.ch, authHeaders)

	reqBody := getUpdatePasswordReqBody(
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "password updated successfully"},
	}

	testUpdatePassword(uat, http.StatusOK, expectedRespData, authHeaders, reqBody)
}

func (uat *userAPITestSuite) TestUpdatePasswordClientAuthenticationFailure() {
	defaultAuthHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(uat.T(), uat.deps.cfg.EventConfig(), uat.deps.cl, uat.deps.ch, defaultAuthHeaders)

	reqBody := getUpdatePasswordReqBody(
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	clientId := defaultAuthHeaders["CLIENT-ID"]
	clientSecret := defaultAuthHeaders["CLIENT-SECRET"]

	expectedRespData := func(msg string) contract.APIResponse {
		return contract.APIResponse{
			Success: false,
			Error:   &contract.Error{Message: msg},
		}
	}

	testCases := map[string]struct {
		authHeader map[string]string
	}{
		"test failure when client id is invalid": {
			authHeader: map[string]string{
				"CLIENT-ID":     "invalid",
				"CLIENT-SECRET": clientSecret,
			},
		},
		"test failure when client secret is invalid": {
			authHeader: map[string]string{
				"CLIENT-ID":     clientId,
				"CLIENT-SECRET": "invalid",
			},
		},
	}

	for name, testCase := range testCases {
		uat.T().Run(name, func(t *testing.T) {
			testUpdatePassword(
				uat,
				http.StatusUnauthorized,
				expectedRespData("authentication failed"),
				testCase.authHeader,
				reqBody,
			)
		})
	}
}

func (uat *userAPITestSuite) TestUpdatePasswordFailure() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl, map[string]interface{}{})
	signUpUser(uat.T(), uat.deps.cfg.EventConfig(), uat.deps.cl, uat.deps.ch, authHeaders)

	expectedRespData := func(msg string) contract.APIResponse {
		return contract.APIResponse{
			Success: false,
			Error:   &contract.Error{Message: msg},
		}
	}

	testCases := map[string]struct {
		data       map[string]interface{}
		statusCode int
		errMsg     string
	}{
		"test failure when email is empty": {
			data:       map[string]interface{}{userEmailKey: EmptyString},
			statusCode: http.StatusBadRequest,
			errMsg:     "email cannot be empty",
		},
		"test failure when email is incorrect": {
			data:       map[string]interface{}{userEmailKey: "other@other.com"},
			statusCode: http.StatusUnauthorized,
			errMsg:     "invalid credentials",
		},
		"test failure when old password is empty": {
			data:       map[string]interface{}{userPasswordKey: EmptyString},
			statusCode: http.StatusBadRequest,
			errMsg:     "old password cannot be empty",
		},
		"test failure when old password is incorrect": {
			data:       map[string]interface{}{userPasswordKey: "invalidPassword"},
			statusCode: http.StatusUnauthorized,
			errMsg:     "invalid credentials",
		},
		"test failure when new password is empty": {
			data:       map[string]interface{}{userNewPasswordKey: EmptyString},
			statusCode: http.StatusBadRequest,
			errMsg:     "new password cannot be empty",
		},
		"test failure when password is below min characters": {
			data:       map[string]interface{}{userNewPasswordKey: RandString(6)},
			statusCode: http.StatusBadRequest,
			errMsg:     "password must be at least 8 characters long",
		},
		"test failure when new password does not match spec": {
			data:       map[string]interface{}{userNewPasswordKey: RandString(8)},
			statusCode: http.StatusBadRequest,
			errMsg:     "password must have at least 1 number, 1 lower character, 1 upper character and 1 symbol",
		},
	}

	for name, testCase := range testCases {
		uat.Run(name, func() {
			testUpdatePassword(
				uat,
				testCase.statusCode,
				expectedRespData(testCase.errMsg),
				authHeaders,
				getUpdatePasswordReqBody(testCase.data),
			)
		})
	}
}

func testSignUpUser(
	t *testing.T,
	cfg config.EventConfig,
	cl *http.Client,
	ch *amqp.Channel,
	expectedCode int,
	expectedRespData contract.APIResponse,
	reqHeaders map[string]string,
	reqBody contract.CreateUserRequest,
	consumeMessage bool,
) {

	b, err := json.Marshal(&reqBody)
	require.NoError(t, err)

	req := newRequest(t, http.MethodPost, "user/sign-up", bytes.NewBuffer(b))
	for key, value := range reqHeaders {
		req.Header.Set(key, value)
	}

	resp := execRequest(t, cl, req)

	responseData := getData(t, expectedCode, resp)

	verifyResp(t, expectedRespData, responseData, true, nil)

	if consumeMessage {
		queueName, ok := cfg.QueueMap()[cfg.SignUpEventCode()]
		if ok {
			testMessageConsume(t, queueName, ch)
		}
	}
}

func testUpdatePassword(
	uat *userAPITestSuite,
	expectedCode int,
	expectedRespData contract.APIResponse,
	reqHeader map[string]string,
	reqBody contract.UpdatePasswordRequest,
) {

	b, err := json.Marshal(&reqBody)
	require.NoError(uat.T(), err)

	req := newRequest(uat.T(), http.MethodPost, "user/update-password", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(uat.T(), uat.deps.cl, req)

	responseData := getData(uat.T(), expectedCode, resp)

	verifyResp(uat.T(), expectedRespData, responseData, true, nil)
}

func getCreateUserReqBody(data map[string]interface{}) contract.CreateUserRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.CreateUserRequest{
		Name:     either(data[userNameKey], RandString(8)).(string),
		Email:    either(data[userEmailKey], NewEmail()).(string),
		Password: either(data[userPasswordKey], NewPassword()).(string),
	}

}

func getUpdatePasswordReqBody(data map[string]interface{}) contract.UpdatePasswordRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.UpdatePasswordRequest{
		Email:       either(data[userEmailKey], NewEmail()).(string),
		OldPassword: either(data[userPasswordKey], NewPassword()).(string),
		NewPassword: either(data[userNewPasswordKey], NewPassword()).(string),
	}

}
