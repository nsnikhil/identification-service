// build integration_test

package test_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/app"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/http/contract"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	userName        = "Test Name"
	userEmail       = "test@test.com"
	userPassword    = "Password@1234"
	newUserPassword = "NewPassword@1234"

	address = "http://127.0.0.1:8080"
)

type componentTestSuite struct {
	cl  *http.Client
	db  *sql.DB
	ch  *amqp.Channel
	cfg config.Config
	suite.Suite
}

func (cst *componentTestSuite) SetupSuite() {
	configFile := "../../local.env"
	startApp(configFile)

	cfg := config.NewConfig(configFile)

	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	require.NoError(cst.T(), err)

	conn, err := amqp.Dial(cfg.AMPQConfig().Address())
	require.NoError(cst.T(), err)

	ch, err := conn.Channel()
	require.NoError(cst.T(), err)

	cst.db = db
	cst.ch = ch
	cst.cfg = cfg
	cst.cl = &http.Client{Timeout: time.Minute}
}

func (cst *componentTestSuite) AfterTest(suiteName, testName string) {
	truncateTables(cst.T(), cst.db)
}

func (cst *componentTestSuite) TestPing() {
	req := newRequest(cst.T(), http.MethodGet, "ping", nil)
	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), http.StatusOK, resp)
	expectedData := contract.APIResponse{Success: true, Data: "pong"}

	verifyData(cst.T(), expectedData, responseData)
}

func (cst *componentTestSuite) TestRegisterClientSuccess() {
	reqBody := contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 10, SessionTTL: 86400}

	testRegisterClient(cst,
		http.StatusCreated,
		contract.APIResponse{Success: true},
		reqBody,
	)
}

func (cst *componentTestSuite) TestRegisterClientFailureForValidation() {
	testCases := map[string]struct {
		reqBody       contract.CreateClientRequest
		expectedError string
	}{
		"test failure when name is empty": {
			reqBody:       contract.CreateClientRequest{Name: "", AccessTokenTTL: 10, SessionTTL: 86400},
			expectedError: "name cannot be empty",
		},
		"test failure when access token ttl is less than 1": {
			reqBody:       contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 0, SessionTTL: 86400},
			expectedError: "access token ttl cannot be less than one",
		},
		"test failure when session ttl is less than 1": {
			reqBody:       contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 10, SessionTTL: 0},
			expectedError: "session ttl cannot be less than one",
		},
		"test failure when session ttl is less than access token ttl": {
			reqBody:       contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 20, SessionTTL: 10},
			expectedError: "session ttl cannot be less than access token ttl",
		},
	}

	for name, testCase := range testCases {
		cst.T().Run(name, func(t *testing.T) {
			testRegisterClient(cst,
				http.StatusBadRequest,
				contract.APIResponse{Success: false, Error: &contract.Error{Message: testCase.expectedError}},
				testCase.reqBody,
			)
		})
	}
}

func (cst *componentTestSuite) TestRegisterClientFailureForDuplicateRecord() {
	reqBody := contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 10, SessionTTL: 86400}

	testRegisterClient(cst,
		http.StatusCreated,
		contract.APIResponse{Success: true},
		reqBody,
	)

	testRegisterClient(cst,
		http.StatusInternalServerError,
		contract.APIResponse{Success: false, Error: &contract.Error{Message: "internal server error"}},
		reqBody,
	)
}

func testRegisterClient(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, reqBody contract.CreateClientRequest) {
	responseData := testCreateClient(cst, expectedCode, reqBody)

	assert.Equal(cst.T(), expectedRespData.Success, responseData.Success)

	if expectedRespData.Success {
		assert.NotNil(cst.T(), responseData.Data.(map[string]interface{})["secret"])
	} else {
		assert.Equal(cst.T(), expectedRespData.Error, responseData.Error)
	}
}

func testCreateClient(cst *componentTestSuite, expectedCode int, reqBody contract.CreateClientRequest) contract.APIResponse {
	b, err := json.Marshal(&reqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "client/register", bytes.NewBuffer(b))
	ac := cst.cfg.AuthConfig()
	req.SetBasicAuth(ac.UserName(), ac.Password())
	resp := execRequest(cst.T(), cst.cl, req)

	return getData(cst.T(), expectedCode, resp)
}

func (cst *componentTestSuite) TestRevokeClientSuccess() {
	regReqBody := contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 10, SessionTTL: 86400}

	testRegisterClient(cst, http.StatusCreated, contract.APIResponse{Success: true}, regReqBody)

	var clientID string
	err := cst.db.QueryRow("select id from clients where name = $1", "clientOne").Scan(&clientID)
	require.NoError(cst.T(), err)
	require.NotEmpty(cst.T(), clientID)

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "client revoked successfully"},
	}

	testRevokeClient(cst, http.StatusOK, expectedRespData, clientID)
}

func (cst *componentTestSuite) TestRevokeClientFailure() {
	expectedRespData := contract.APIResponse{
		Success: false,
		Error: &contract.Error{
			Message: "resource not found",
		},
	}

	testRevokeClient(cst, http.StatusNotFound, expectedRespData, "86d690dd-92a0-40ac-ad48-110c951e3cb8")
}

func testRevokeClient(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, clientID string) {
	revReqBody := contract.ClientRevokeRequest{ID: clientID}

	b, err := json.Marshal(&revReqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "client/revoke", bytes.NewBuffer(b))
	ac := cst.cfg.AuthConfig()
	req.SetBasicAuth(ac.UserName(), ac.Password())
	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), expectedCode, resp)

	assert.Equal(cst.T(), expectedRespData.Success, responseData.Success)

	if expectedRespData.Success {
		assert.Equal(cst.T(), expectedRespData.Data, responseData.Data)
	} else {
		assert.Equal(cst.T(), expectedRespData.Error, responseData.Error)
	}

}

func createAndGetClientCredentials(cst *componentTestSuite) map[string]string {
	clientReqBody := contract.CreateClientRequest{Name: "clientOne", AccessTokenTTL: 10, SessionTTL: 87601}
	clientResp := testCreateClient(cst, http.StatusCreated, clientReqBody)
	require.True(cst.T(), clientResp.Success)

	return map[string]string{
		"CLIENT-ID":     "clientOne",
		"CLIENT-SECRET": clientResp.Data.(map[string]interface{})["secret"].(string),
	}
}

func createUser(cst *componentTestSuite, consumeMessage bool) map[string]string {
	headers := createAndGetClientCredentials(cst)

	reqBody := contract.CreateUserRequest{Name: userName, Email: userEmail, Password: userPassword}

	expectedRespData := contract.APIResponse{Success: true, Data: map[string]interface{}{"message": "user created successfully"}}

	testSignUpUser(cst, http.StatusCreated, expectedRespData, headers, reqBody)
	if consumeMessage {
		testMessageConsume(cst.T(), cst.cfg.AMPQConfig(), cst.ch)
	}

	return headers
}

func (cst *componentTestSuite) TestSignUpUserSuccess() {
	createUser(cst, true)
}

func (cst *componentTestSuite) TestSignUpUserValidationFailure() {
	headers := createAndGetClientCredentials(cst)

	testCases := map[string]struct {
		reqBody      contract.CreateUserRequest
		errorMessage string
	}{
		"test failure when name is empty": {
			reqBody:      contract.CreateUserRequest{Name: "", Email: userEmail, Password: userPassword},
			errorMessage: "name cannot be empty",
		},
		"test failure when email is empty": {
			reqBody:      contract.CreateUserRequest{Name: userName, Email: "", Password: userPassword},
			errorMessage: "email cannot be empty",
		},
		"test failure when password is empty": {
			reqBody:      contract.CreateUserRequest{Name: userName, Email: userEmail, Password: ""},
			errorMessage: "password cannot be empty",
		},
		"test failure when password is invalid": {
			reqBody:      contract.CreateUserRequest{Name: userName, Email: userEmail, Password: "invalid"},
			errorMessage: "password must be at least 8 characters long",
		},
	}

	for name, testCase := range testCases {
		cst.T().Run(name, func(t *testing.T) {
			testSignUpUser(cst,
				http.StatusBadRequest,
				contract.APIResponse{Success: false, Error: &contract.Error{Message: testCase.errorMessage}},
				headers,
				testCase.reqBody,
			)
		})
	}
}

func (cst *componentTestSuite) TestSignUpUserFailureForDuplicateRecord() {
	headers := createAndGetClientCredentials(cst)

	reqBody := contract.CreateUserRequest{Name: userName, Email: userEmail, Password: userPassword}

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "user created successfully"},
	}

	testSignUpUser(cst, http.StatusCreated, expectedRespData, headers, reqBody)

	expectedRespData = contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "internal server error"},
	}

	testSignUpUser(cst, http.StatusInternalServerError, expectedRespData, headers, reqBody)
}

func testSignUpUser(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, reqHeaders map[string]string, reqBody contract.CreateUserRequest) {
	b, err := json.Marshal(&reqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "user/sign-up", bytes.NewBuffer(b))
	for key, value := range reqHeaders {
		req.Header.Set(key, value)
	}

	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), expectedCode, resp)

	assert.Equal(cst.T(), expectedRespData, responseData)
}

func (cst *componentTestSuite) TestUpdatePasswordSuccess() {
	headers := createUser(cst, false)

	reqBody := contract.UpdatePasswordRequest{
		Email:       userEmail,
		OldPassword: userPassword,
		NewPassword: newUserPassword,
	}

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "password updated successfully"},
	}

	testUpdatePassword(cst, http.StatusOK, expectedRespData, headers, reqBody)
}

func (cst *componentTestSuite) TestUpdatePasswordFailure() {
	headers := createUser(cst, false)

	testCases := map[string]struct {
		reqBody    contract.UpdatePasswordRequest
		statusCode int
		errMsg     string
	}{
		"test failure when email is incorrect": {
			reqBody: contract.UpdatePasswordRequest{
				Email:       "other@email.com",
				OldPassword: userPassword,
				NewPassword: newUserPassword,
			},
			statusCode: http.StatusUnauthorized,
			errMsg:     "invalid credentials",
		},
		"test failure when password is incorrect": {
			reqBody: contract.UpdatePasswordRequest{
				Email:       userEmail,
				OldPassword: "InvalidPassword",
				NewPassword: newUserPassword,
			},
			statusCode: http.StatusUnauthorized,
			errMsg:     "invalid credentials",
		},
		"test failure when new password does not match spec": {
			reqBody: contract.UpdatePasswordRequest{
				Email:       userEmail,
				OldPassword: userPassword,
				NewPassword: "invalid",
			},
			statusCode: http.StatusBadRequest,
			errMsg:     "password must be at least 8 characters long",
		},
	}

	for name, testCase := range testCases {
		cst.Run(name, func() {
			testUpdatePassword(cst,
				testCase.statusCode,
				contract.APIResponse{Success: false, Error: &contract.Error{Message: testCase.errMsg}},
				headers,
				testCase.reqBody,
			)
		})
	}
}

func testUpdatePassword(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, reqHeader map[string]string, reqBody contract.UpdatePasswordRequest) {
	b, err := json.Marshal(&reqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "user/update-password", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), expectedCode, resp)

	assert.Equal(cst.T(), expectedRespData.Success, responseData.Success)

	if expectedRespData.Success {
		assert.Equal(cst.T(), expectedRespData.Data, responseData.Data)
	} else {
		assert.Equal(cst.T(), expectedRespData.Error, responseData.Error)
	}
}

func (cst *componentTestSuite) TestCreateSessionSuccess() {
	testCreateSession(cst)
}

func testCreateSession(cst *componentTestSuite) map[string]string {
	header := createUser(cst, false)

	reqBody := contract.LoginRequest{Email: userEmail, Password: userPassword}

	expectedRespData := contract.APIResponse{Success: true}

	testCreateSessionSuccess(cst, http.StatusCreated, expectedRespData, header, reqBody)

	return header
}

func (cst *componentTestSuite) TestCreateSessionFailureWhenCredentialAreIncorrect() {
	header := createUser(cst, false)

	testCases := map[string]contract.LoginRequest{
		"test failure when email is incorrect": {
			Email: "other@email.com", Password: userPassword,
		},
		"test failure when password is incorrect": {
			Email: userEmail, Password: "OtherPassword@1234",
		},
	}

	for name, reqBody := range testCases {
		cst.Run(name, func() {
			expectedRespData := contract.APIResponse{
				Success: false,
				Error:   &contract.Error{Message: "invalid credentials"},
			}

			testCreateSessionSuccess(cst, http.StatusUnauthorized, expectedRespData, header, reqBody)
		})
	}

}

func testCreateSessionSuccess(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, reqHeader map[string]string, reqBody contract.LoginRequest) {
	b, err := json.Marshal(&reqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "session/login", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), expectedCode, resp)

	assert.Equal(cst.T(), expectedRespData.Success, responseData.Success)

	if expectedRespData.Success {
		verifyTokens(cst.T(), responseData.Data)
	} else {
		assert.Equal(cst.T(), expectedRespData.Error, responseData.Error)
	}
}

func (cst *componentTestSuite) TestRefreshTokenSuccess() {
	header := testCreateSession(cst)

	var userID string
	err := cst.db.QueryRow("select id from users where email = $1", userEmail).Scan(&userID)
	require.NoError(cst.T(), err)

	var refreshToken string
	err = cst.db.QueryRow("select refreshtoken from sessions where userid = $1", userID).Scan(&refreshToken)
	require.NoError(cst.T(), err)

	reqBody := contract.RefreshTokenRequest{RefreshToken: refreshToken}

	expectedRespData := contract.APIResponse{
		Success: true,
	}

	testRefreshToken(cst, http.StatusOK, expectedRespData, header, reqBody)
}

func (cst *componentTestSuite) TestRefreshTokenFailureWhenRefreshTokenIsIncorrect() {
	header := testCreateSession(cst)

	reqBody := contract.RefreshTokenRequest{RefreshToken: "28751ed1-9abf-40b4-bfff-2161b450ff39"}

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "internal server error"},
	}

	testRefreshToken(cst, http.StatusInternalServerError, expectedRespData, header, reqBody)
}

func (cst *componentTestSuite) TestRefreshTokenFailureWhenSessionExpires() {
	header := testCreateSession(cst)

	var userID string
	err := cst.db.QueryRow("select id from users where email = $1", userEmail).Scan(&userID)
	require.NoError(cst.T(), err)

	prev := time.Now().AddDate(0, -2, -5)
	_, err = cst.db.Exec("update sessions set createdat=$1, updatedat=$2", prev, prev)
	require.NoError(cst.T(), err)

	var refreshToken string
	err = cst.db.QueryRow("select refreshtoken from sessions where userid = $1", userID).Scan(&refreshToken)
	require.NoError(cst.T(), err)

	reqBody := contract.RefreshTokenRequest{RefreshToken: refreshToken}

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "internal server error"},
	}

	testRefreshToken(cst, http.StatusInternalServerError, expectedRespData, header, reqBody)
}

func testRefreshToken(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, reqHeader map[string]string, reqBody contract.RefreshTokenRequest) {
	b, err := json.Marshal(&reqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "session/refresh-token", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), expectedCode, resp)

	assert.Equal(cst.T(), expectedRespData.Success, responseData.Success)

	if expectedRespData.Success {
		assert.NotEmpty(cst.T(), responseData.Data.(map[string]interface{})["access_token"].(string))
	} else {
		assert.Equal(cst.T(), expectedRespData.Error, responseData.Error)
	}
}

func (cst *componentTestSuite) TestLogoutUserSuccess() {
	header := testCreateSession(cst)

	var userID string
	err := cst.db.QueryRow("select id from users where email = $1", userEmail).Scan(&userID)
	require.NoError(cst.T(), err)

	var refreshToken string
	err = cst.db.QueryRow("select refreshtoken from sessions where userid = $1", userID).Scan(&refreshToken)
	require.NoError(cst.T(), err)

	reqBody := contract.LogoutRequest{RefreshToken: refreshToken}

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "Logout Successful"},
	}

	testLogoutUser(cst, http.StatusOK, expectedRespData, header, reqBody)
}

func (cst *componentTestSuite) TestLogoutUserFailureForIncorrectRefreshToken() {
	header := testCreateSession(cst)

	reqBody := contract.LogoutRequest{RefreshToken: "28751ed1-9abf-40b4-bfff-2161b450ff39"}

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "internal server error"},
	}

	testLogoutUser(cst, http.StatusInternalServerError, expectedRespData, header, reqBody)
}

func testLogoutUser(cst *componentTestSuite, expectedCode int, expectedRespData contract.APIResponse, reqHeader map[string]string, reqBody contract.LogoutRequest) {
	b, err := json.Marshal(&reqBody)
	require.NoError(cst.T(), err)

	req := newRequest(cst.T(), http.MethodPost, "session/logout", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(cst.T(), cst.cl, req)

	responseData := getData(cst.T(), expectedCode, resp)

	assert.Equal(cst.T(), expectedRespData.Success, responseData.Success)

	if expectedRespData.Success {
		assert.Equal(cst.T(), expectedRespData.Data, responseData.Data)
	} else {
		assert.Equal(cst.T(), expectedRespData.Error, responseData.Error)
	}
}

func TestComponents(t *testing.T) {
	suite.Run(t, new(componentTestSuite))
}

func execRequest(t *testing.T, cl *http.Client, req *http.Request) *http.Response {
	resp, err := cl.Do(req)
	require.NoError(t, err)

	return resp
}

func getData(t *testing.T, expectedCode int, resp *http.Response) contract.APIResponse {
	require.Equal(t, expectedCode, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var res contract.APIResponse
	err = json.Unmarshal(b, &res)
	require.NoError(t, err)

	return res
}

func verifyData(t *testing.T, expectedResponse contract.APIResponse, response contract.APIResponse) {
	require.Equal(t, expectedResponse, response)
}

func verifyTokens(t *testing.T, data interface{}) {
	require.NotNil(t, data)

	accessToken := data.(map[string]interface{})["access_token"].(string)
	refreshToken := data.(map[string]interface{})["refresh_token"].(string)

	assert.True(t, len(accessToken) != 0)
	assert.True(t, len(refreshToken) != 0)
}

func newRequest(t *testing.T, method, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", address, path), body)
	require.NoError(t, err)

	return req
}

func startApp(configFile string) {
	go app.StartHTTPServer(configFile)
	time.Sleep(time.Second)
}

func testMessageConsume(t *testing.T, cfg config.AMPQConfig, ch *amqp.Channel) {
	delivery, err := ch.Consume(cfg.QueueName(), "component-test-consumer", true, true, false, false, nil)
	require.NoError(t, err)

	for {
		select {
		case d := <-delivery:
			assert.NotEmpty(t, string(d.Body))
			return
		case <-time.After(time.Second * 2):
			t.Fail()
			return
		}
	}
}

func truncateTables(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`truncate clients`)
	require.NoError(t, err)

	_, err = db.Exec(`truncate users cascade`)
	require.NoError(t, err)

	_, err = db.Exec(`truncate sessions`)
	require.NoError(t, err)
}
