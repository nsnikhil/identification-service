// build component_test

package test

import (
	"bytes"
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/event"
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

func (uat *userAPITestSuite) AfterTest(suiteName, testName string) {
	truncateTables(uat.T(), uat.deps.ctx, uat.deps.db, ClientTableName, UserTableName)
}

func (uat *userAPITestSuite) TearDownSuite() {
	tearDownTest(uat.T(), uat.deps)
}

func TestUserAPI(t *testing.T) {
	suite.Run(t, new(userAPITestSuite))
}

func (uat *userAPITestSuite) TestSignUpUserSuccess() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl)

	reqBody := getCreateUserReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "user created successfully"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.PublisherConfig(),
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
		uat.deps.cfg.PublisherConfig(),
		uat.deps.cl,
		uat.deps.ch,
		http.StatusUnauthorized,
		expectedRespData, map[string]string{},
		reqBody,
		false,
	)
}

func (uat *userAPITestSuite) TestSignUpUserValidationFailure() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl)

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
		"test failure when password is invalid": {
			data:         map[string]interface{}{userPasswordKey: UserPasswordInvalid},
			errorMessage: "password must have at least 1 number, 1 lower character, 1 upper character and 1 symbol",
		},
	}

	for name, testCase := range testCases {
		uat.T().Run(name, func(t *testing.T) {
			testSignUpUser(
				uat.T(),
				uat.deps.cfg.PublisherConfig(),
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

//TODO: WHY DOES DUPLICATE RECORD RETURN INTERNAL SERVER ERROR ?
func (uat *userAPITestSuite) TestSignUpUserFailureForDuplicateRecord() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl)

	reqBody := getCreateUserReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "user created successfully"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.PublisherConfig(),
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
		Error:   &contract.Error{Message: "internal server error"},
	}

	testSignUpUser(
		uat.T(),
		uat.deps.cfg.PublisherConfig(),
		uat.deps.cl,
		uat.deps.ch,
		http.StatusInternalServerError,
		expectedRespData,
		authHeaders,
		reqBody,
		false,
	)
}

func (uat *userAPITestSuite) TestUpdatePasswordSuccess() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl)
	signUpUser(uat.T(), uat.deps.cfg.PublisherConfig(), uat.deps.cl, uat.deps.ch, authHeaders)

	reqBody := getUpdatePasswordReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "password updated successfully"},
	}

	testUpdatePassword(uat, http.StatusOK, expectedRespData, authHeaders, reqBody)
}

func (uat *userAPITestSuite) TestUpdatePasswordFailure() {
	authHeaders := registerClientAndGetHeaders(uat.T(), uat.deps.cfg.AuthConfig(), uat.deps.cl)
	signUpUser(uat.T(), uat.deps.cfg.PublisherConfig(), uat.deps.cl, uat.deps.ch, authHeaders)

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
		"test failure when email is incorrect": {
			data:       map[string]interface{}{userEmailKey: "other@other.com"},
			statusCode: http.StatusUnauthorized,
			errMsg:     "invalid credentials",
		},
		"test failure when password is incorrect": {
			data:       map[string]interface{}{userPasswordKey: "invalidPassword"},
			statusCode: http.StatusUnauthorized,
			errMsg:     "invalid credentials",
		},
		"test failure when new password does not match spec": {
			data:       map[string]interface{}{userNewPasswordKey: UserPasswordInvalid},
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
	cfg config.PublisherConfig,
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
		queueName, ok := cfg.QueueMap()[string(event.SignUp)]
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
		Name:     either(data[userNameKey], UserName).(string),
		Email:    either(data[userEmailKey], UserEmail).(string),
		Password: either(data[userPasswordKey], UserPassword).(string),
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
		Email:       either(data[userEmailKey], UserEmail).(string),
		OldPassword: either(data[userPasswordKey], UserPassword).(string),
		NewPassword: either(data[userNewPasswordKey], UserPasswordNew).(string),
	}

}
