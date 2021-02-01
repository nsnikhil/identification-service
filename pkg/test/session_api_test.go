package test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/http/contract"
	"net/http"
	"testing"
	"time"
)

const (
	sessionRefreshTokenKey = "refreshToken"
)

type sessionAPITestSuite struct {
	deps testDeps
	suite.Suite
}

func (sat *sessionAPITestSuite) SetupSuite() {
	sat.deps = setupTest(sat.T())
}

func (sat *sessionAPITestSuite) TearDownSuite() {
	truncateTables(sat.T(), sat.deps.ctx, sat.deps.db, ClientTableName, UserTableName, SessionTableName)
	tearDownTest(sat.T(), sat.deps)
}

func TestSessionAPI(t *testing.T) {
	suite.Run(t, new(sessionAPITestSuite))
}

func (sat *sessionAPITestSuite) TestLoginSuccess() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)

	reqBody := getLoginReqBody(
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	testLogin(sat.T(), sat.deps.cl, http.StatusCreated, contract.APIResponse{Success: true}, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestLoginFailureWhenClientCredentialsAreMissing() {
	reqBody := getLoginReqBody(map[string]interface{}{})

	expectedResp := contract.APIResponse{
		Success: false,
		Error: &contract.Error{
			Message: "authentication failed",
		},
	}

	testLogin(sat.T(), sat.deps.cl, http.StatusUnauthorized, expectedResp, map[string]string{}, reqBody)
}

func (sat *sessionAPITestSuite) TestLoginFailedWhenClientAuthenticationFailed() {
	defaultAuthHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	clientId := defaultAuthHeaders["CLIENT-ID"]
	clientSecret := defaultAuthHeaders["CLIENT-SECRET"]

	expectedResp := contract.APIResponse{
		Success: false,
		Error: &contract.Error{
			Message: "authentication failed",
		},
	}

	reqBody := getLoginReqBody(map[string]interface{}{})

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
		sat.Run(name, func() {
			testLogin(sat.T(), sat.deps.cl, http.StatusUnauthorized, expectedResp, testCase.authHeader, reqBody)
		})
	}
}

func (sat *sessionAPITestSuite) TestLoginSuccessWithRevokeOld() {
	maxActiveSessions := 2
	clientData := map[string]interface{}{
		clientMaxActiveSessionsKey: maxActiveSessions,
	}

	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, clientData)
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)

	reqBody := getLoginReqBody(
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	expectedResp := contract.APIResponse{Success: true}

	testLogin(sat.T(), sat.deps.cl, http.StatusCreated, expectedResp, authHeaders, reqBody)
	testLogin(sat.T(), sat.deps.cl, http.StatusCreated, expectedResp, authHeaders, reqBody)
	testLogin(sat.T(), sat.deps.cl, http.StatusCreated, expectedResp, authHeaders, reqBody)

	var userID string

	err := sat.deps.db.QueryRowContext(
		context.Background(),
		`select id from users where email=$1`,
		userDetails.Email,
	).Scan(&userID)

	require.NoError(sat.T(), err)
	assert.NotEmpty(sat.T(), userID)

	var revokedCount int
	err = sat.deps.db.QueryRowContext(
		context.Background(),
		`select count(*) from sessions where user_id=$1 and revoked=true`,
		userID,
	).Scan(&revokedCount)

	require.NoError(sat.T(), err)
	assert.Equal(sat.T(), 1, revokedCount)
}

func (sat *sessionAPITestSuite) TestLoginFailureWhenValidationFails() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	expectedRespData := func(msg string) contract.APIResponse {
		return contract.APIResponse{
			Success: false,
			Error:   &contract.Error{Message: msg},
		}
	}

	testCases := map[string]struct {
		data   map[string]interface{}
		errMsg string
	}{
		"test failure when email is empty": {
			data:   map[string]interface{}{userEmailKey: EmptyString},
			errMsg: "email cannot be empty",
		},
		"test failure when password is empty": {
			data:   map[string]interface{}{userPasswordKey: EmptyString},
			errMsg: "password cannot be empty",
		},
	}

	for name, testCase := range testCases {
		sat.Run(name, func() {
			testLogin(
				sat.T(),
				sat.deps.cl,
				http.StatusBadRequest,
				expectedRespData(testCase.errMsg),
				authHeaders,
				getLoginReqBody(testCase.data),
			)
		})
	}

}

func (sat *sessionAPITestSuite) TestLoginFailureWhenCredentialsAreIncorrect() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "invalid credentials"},
	}

	testCases := map[string]struct {
		data   map[string]interface{}
		errMsg string
	}{
		"test failure when email is incorrect": {
			data: map[string]interface{}{
				userEmailKey:    "other@other.com",
				userPasswordKey: userDetails.Password,
			},
		},
		"test failure when password is incorrect": {
			data: map[string]interface{}{
				userEmailKey:    userDetails.Email,
				userPasswordKey: "invalidPassword",
			},
		},
	}

	for name, testCase := range testCases {
		sat.Run(name, func() {
			testLogin(
				sat.T(),
				sat.deps.cl,
				http.StatusUnauthorized,
				expectedRespData,
				authHeaders,
				getLoginReqBody(testCase.data),
			)
		})
	}

}

func (sat *sessionAPITestSuite) TestRefreshTokenSuccess() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)
	loginUser(sat.T(), sat.deps.cl, authHeaders,
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	var userID string

	err := sat.deps.db.QueryRowContext(
		sat.deps.ctx,
		"select id from users where email = $1",
		userDetails.Email,
	).Scan(&userID)

	require.NoError(sat.T(), err)
	require.NotEmpty(sat.T(), userID)

	var refreshToken string

	err = sat.deps.db.QueryRowContext(
		sat.deps.ctx,
		"select refresh_token from sessions where user_id = $1",
		userID,
	).Scan(&refreshToken)

	require.NoError(sat.T(), err)
	require.NotEmpty(sat.T(), refreshToken)

	reqBody := getRefreshTokenReqBody(map[string]interface{}{sessionRefreshTokenKey: refreshToken})
	expectedRespData := contract.APIResponse{Success: true}

	testRefreshToken(sat, http.StatusOK, expectedRespData, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestRefreshTokenWhenClientAuthenticationFailed() {
	defaultAuthHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	clientId := defaultAuthHeaders["CLIENT-ID"]
	clientSecret := defaultAuthHeaders["CLIENT-SECRET"]

	expectedResp := contract.APIResponse{
		Success: false,
		Error: &contract.Error{
			Message: "authentication failed",
		},
	}

	reqBody := getRefreshTokenReqBody(map[string]interface{}{})

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
		sat.Run(name, func() {
			testRefreshToken(sat, http.StatusUnauthorized, expectedResp, testCase.authHeader, reqBody)
		})
	}
}

func (sat *sessionAPITestSuite) TestRefreshTokenFailureWhenClientCredentialsAreMissing() {
	reqBody := getRefreshTokenReqBody(map[string]interface{}{sessionRefreshTokenKey: NewUUID()})

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "authentication failed"},
	}

	testRefreshToken(sat, http.StatusUnauthorized, expectedRespData, map[string]string{}, reqBody)
}

func (sat *sessionAPITestSuite) TestRefreshTokenFailureWhenValidationFails() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	reqBody := contract.RefreshTokenRequest{RefreshToken: EmptyString}

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "refresh token cannot be empty"},
	}

	testRefreshToken(sat, http.StatusBadRequest, expectedRespData, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestRefreshTokenFailureWhenRefreshTokenIsIncorrect() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)
	loginUser(sat.T(), sat.deps.cl, authHeaders,
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	reqBody := contract.RefreshTokenRequest{RefreshToken: NewUUID()}

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "resource not found"},
	}

	testRefreshToken(sat, http.StatusNotFound, expectedRespData, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestRefreshTokenFailureWhenSessionExpires() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)
	loginUser(sat.T(), sat.deps.cl, authHeaders,
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	var userID string
	err := sat.deps.db.QueryRowContext(
		sat.deps.ctx,
		"select id from users where email = $1",
		userDetails.Email,
	).Scan(&userID)

	require.NoError(sat.T(), err)
	require.NotEmpty(sat.T(), userID)

	prev := time.Now().AddDate(0, -2, -5)
	_, err = sat.deps.db.ExecContext(
		sat.deps.ctx,
		"update sessions set created_at=$1, updated_at=$2",
		prev,
		prev,
	)

	require.NoError(sat.T(), err)

	var refreshToken string
	err = sat.deps.db.QueryRowContext(
		sat.deps.ctx,
		"select refresh_token from sessions where user_id = $1",
		userID,
	).Scan(&refreshToken)

	require.NoError(sat.T(), err)
	require.NotEmpty(sat.T(), refreshToken)

	reqBody := contract.RefreshTokenRequest{RefreshToken: refreshToken}

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "authentication failed"},
	}

	testRefreshToken(sat, http.StatusUnauthorized, expectedRespData, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestLogoutUserSuccess() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})
	userDetails := signUpUser(sat.T(), sat.deps.cfg.EventConfig(), sat.deps.cl, sat.deps.ch, authHeaders)
	loginUser(sat.T(), sat.deps.cl, authHeaders,
		map[string]interface{}{
			userEmailKey:    userDetails.Email,
			userPasswordKey: userDetails.Password,
		},
	)

	var userID string

	err := sat.deps.db.QueryRowContext(
		sat.deps.ctx,
		"select id from users where email = $1",
		userDetails.Email,
	).Scan(&userID)

	require.NoError(sat.T(), err)
	require.NotEmpty(sat.T(), userID)

	var refreshToken string

	err = sat.deps.db.QueryRowContext(
		sat.deps.ctx,
		"select refresh_token from sessions where user_id = $1",
		userID,
	).Scan(&refreshToken)

	require.NoError(sat.T(), err)
	require.NotEmpty(sat.T(), refreshToken)

	reqBody := getLogoutReqBody(map[string]interface{}{sessionRefreshTokenKey: refreshToken})

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "Logout Successful"},
	}

	testLogoutUser(sat, http.StatusOK, expectedRespData, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestLogoutFailureWhenClientCredentialsAreMissing() {
	reqBody := getLogoutReqBody(map[string]interface{}{})

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "authentication failed"},
	}

	testLogoutUser(sat, http.StatusUnauthorized, expectedRespData, map[string]string{}, reqBody)
}

func (sat *sessionAPITestSuite) TestLogoutFailureWhenClientAuthenticationFailed() {
	defaultAuthHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	clientId := defaultAuthHeaders["CLIENT-ID"]
	clientSecret := defaultAuthHeaders["CLIENT-SECRET"]

	expectedResp := contract.APIResponse{
		Success: false,
		Error: &contract.Error{
			Message: "authentication failed",
		},
	}

	reqBody := getLogoutReqBody(map[string]interface{}{})

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
		sat.Run(name, func() {
			testLogoutUser(sat, http.StatusUnauthorized, expectedResp, testCase.authHeader, reqBody)
		})
	}
}

func (sat *sessionAPITestSuite) TestLogoutUserFailureForIncorrectRefreshToken() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	reqBody := getLogoutReqBody(map[string]interface{}{sessionRefreshTokenKey: NewUUID()})

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "resource not found"},
	}

	testLogoutUser(sat, http.StatusNotFound, expectedRespData, authHeaders, reqBody)
}

func (sat *sessionAPITestSuite) TestLogoutUserFailureForValidationFailure() {
	authHeaders := registerClientAndGetHeaders(sat.T(), sat.deps.cfg.AuthConfig(), sat.deps.cl, map[string]interface{}{})

	reqBody := getLogoutReqBody(map[string]interface{}{sessionRefreshTokenKey: EmptyString})

	expectedRespData := contract.APIResponse{
		Success: false,
		Error:   &contract.Error{Message: "refresh token cannot be empty"},
	}

	testLogoutUser(sat, http.StatusBadRequest, expectedRespData, authHeaders, reqBody)
}

func testLogin(
	t *testing.T,
	cl *http.Client,
	expectedCode int,
	expectedRespData contract.APIResponse,
	reqHeader map[string]string,
	reqBody contract.LoginRequest,
) {

	b, err := json.Marshal(&reqBody)
	require.NoError(t, err)

	req := newRequest(t, http.MethodPost, "session/login", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(t, cl, req)

	responseData := getData(t, expectedCode, resp)

	verifyResp(t, expectedRespData, responseData, false, func(data interface{}) []interface{} {
		var res []interface{}
		res = append(res, data.(map[string]interface{})["access_token"])
		res = append(res, data.(map[string]interface{})["refresh_token"])
		return res
	})
}

func testRefreshToken(
	sat *sessionAPITestSuite,
	expectedCode int,
	expectedRespData contract.APIResponse,
	reqHeader map[string]string,
	reqBody contract.RefreshTokenRequest,
) {

	b, err := json.Marshal(&reqBody)
	require.NoError(sat.T(), err)

	req := newRequest(sat.T(), http.MethodPost, "session/refresh-token", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(sat.T(), sat.deps.cl, req)

	responseData := getData(sat.T(), expectedCode, resp)

	verifyResp(sat.T(), expectedRespData, responseData, false, func(data interface{}) []interface{} {
		var res []interface{}
		res = append(res, data.(map[string]interface{})["access_token"])
		return res
	})
}

func testLogoutUser(
	sat *sessionAPITestSuite,
	expectedCode int,
	expectedRespData contract.APIResponse,
	reqHeader map[string]string,
	reqBody contract.LogoutRequest,
) {

	b, err := json.Marshal(&reqBody)
	require.NoError(sat.T(), err)

	req := newRequest(sat.T(), http.MethodPost, "session/logout", bytes.NewBuffer(b))
	for key, value := range reqHeader {
		req.Header.Set(key, value)
	}

	resp := execRequest(sat.T(), sat.deps.cl, req)

	responseData := getData(sat.T(), expectedCode, resp)

	verifyResp(sat.T(), expectedRespData, responseData, true, nil)
}

func getLoginReqBody(data map[string]interface{}) contract.LoginRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.LoginRequest{
		Email:    either(data[userEmailKey], NewEmail()).(string),
		Password: either(data[userPasswordKey], NewPassword()).(string),
	}
}

func getRefreshTokenReqBody(data map[string]interface{}) contract.RefreshTokenRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.RefreshTokenRequest{
		RefreshToken: either(data[sessionRefreshTokenKey], NewUUID()).(string),
	}
}

func getLogoutReqBody(data map[string]interface{}) contract.LogoutRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.LogoutRequest{
		RefreshToken: either(data[sessionRefreshTokenKey], NewUUID()).(string),
	}
}
