package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/http/contract"
	"net/http"
	"testing"
)

const (
	clientNameKey                = "name"
	clientAccessTokenTTLKey      = "accessTokenTTL"
	clientSessionTTLKey          = "sessionTTL"
	clientMaxActiveSessionsKey   = "maxActiveSessions"
	clientSessionStrategyNameKey = "sessionStrategyName"
)

type clientAPITestSuite struct {
	deps testDeps
	suite.Suite
}

func (cat *clientAPITestSuite) SetupSuite() {
	cat.deps = setupTest(cat.T())
}

func (cat *clientAPITestSuite) TearDownSuite() {
	truncateTables(cat.T(), cat.deps.ctx, cat.deps.db, ClientTableName)
	tearDownTest(cat.T(), cat.deps)
}

func TestClientAPI(t *testing.T) {
	suite.Run(t, new(clientAPITestSuite))
}

func (cat *clientAPITestSuite) TestRegisterClientSuccess() {
	reqBody := getRegisterClientReqBody(map[string]interface{}{})

	testRegisterClient(
		cat.T(),
		cat.deps.cfg.AuthConfig(),
		cat.deps.cl,
		http.StatusCreated,
		contract.APIResponse{Success: true},
		reqBody,
	)
}

func (cat *clientAPITestSuite) TestRegisterClientValidationFailure() {
	randomString := RandString(8)

	testCases := map[string]struct {
		data          map[string]interface{}
		expectedError string
	}{
		"test failure when name is empty": {
			data:          map[string]interface{}{clientNameKey: EmptyString},
			expectedError: "name cannot be empty",
		},
		"test failure when access token ttl is less than 1": {
			data:          map[string]interface{}{clientAccessTokenTTLKey: Zero},
			expectedError: "access token ttl cannot be less than one",
		},
		"test failure when session ttl is less than 1": {
			data:          map[string]interface{}{clientSessionTTLKey: Zero},
			expectedError: "session ttl cannot be less than one",
		},
		"test failure when session ttl is less than access token ttl": {
			data:          map[string]interface{}{clientSessionTTLKey: 10, clientAccessTokenTTLKey: 20},
			expectedError: "session ttl cannot be less than access token ttl",
		},
		"test failure when max active session is less than 1": {
			data:          map[string]interface{}{clientMaxActiveSessionsKey: Zero},
			expectedError: "max active sessions cannot be less than one",
		},
		"test failure when session strategy is empty": {
			data:          map[string]interface{}{clientSessionStrategyNameKey: EmptyString},
			expectedError: "session strategy name cannot be empty",
		},
		"test failure when session strategy is invalid": {
			data:          map[string]interface{}{clientSessionStrategyNameKey: randomString},
			expectedError: fmt.Sprintf("invalid session strategy %s", randomString),
		},
	}

	for name, testCase := range testCases {
		cat.T().Run(name, func(t *testing.T) {
			testRegisterClient(cat.T(),
				cat.deps.cfg.AuthConfig(),
				cat.deps.cl,
				http.StatusBadRequest,
				contract.APIResponse{
					Success: false,
					Error:   &contract.Error{Message: testCase.expectedError},
				},
				getRegisterClientReqBody(testCase.data),
			)
		})
	}
}

func (cat *clientAPITestSuite) TestRegisterClientFailureForDuplicateRecord() {
	reqBody := getRegisterClientReqBody(map[string]interface{}{})
	testRegisterClient(
		cat.T(),
		cat.deps.cfg.AuthConfig(),
		cat.deps.cl,
		http.StatusCreated,
		contract.APIResponse{Success: true},
		reqBody,
	)

	testRegisterClient(
		cat.T(),
		cat.deps.cfg.AuthConfig(),
		cat.deps.cl,
		http.StatusConflict,
		contract.APIResponse{
			Success: false, Error: &contract.Error{Message: "duplicate record"},
		},
		reqBody,
	)
}

func (cat *clientAPITestSuite) TestRevokeClientSuccess() {
	authHeaders := registerClientAndGetHeaders(cat.T(), cat.deps.cfg.AuthConfig(), cat.deps.cl, nil)

	var clientID string

	err := cat.deps.db.QueryRowContext(
		cat.deps.ctx, `select id from clients where name = $1`, authHeaders["CLIENT-ID"],
	).Scan(&clientID)

	require.NoError(cat.T(), err)
	require.NotEmpty(cat.T(), clientID)

	expectedRespData := contract.APIResponse{
		Success: true,
		Data:    map[string]interface{}{"message": "client revoked successfully"},
	}

	testRevokeClient(cat, http.StatusOK, expectedRespData, clientID)
}

func (cat *clientAPITestSuite) TestRevokeClientFailure() {
	expectedRespData := contract.APIResponse{
		Success: false,
		Error: &contract.Error{
			Message: "resource not found",
		},
	}

	testRevokeClient(cat, http.StatusNotFound, expectedRespData, NewUUID())
}

func getRegisterClientReqBody(data map[string]interface{}) contract.CreateClientRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.CreateClientRequest{
		Name:              either(data[clientNameKey], RandString(8)).(string),
		AccessTokenTTL:    either(data[clientAccessTokenTTLKey], RandInt(1, 10)).(int),
		SessionTTL:        either(data[clientSessionTTLKey], RandInt(1440, 86701)).(int),
		MaxActiveSessions: either(data[clientMaxActiveSessionsKey], RandInt(1, 10)).(int),
		SessionStrategy:   either(data[clientSessionStrategyNameKey], ClientSessionStrategyRevokeOld).(string),
	}
}

func testRegisterClient(
	t *testing.T,
	cfg config.AuthConfig,
	cl *http.Client,
	expectedCode int,
	expectedRespData contract.APIResponse,
	reqBody contract.CreateClientRequest,
) contract.APIResponse {

	b, err := json.Marshal(&reqBody)
	require.NoError(t, err)

	req := newRequest(t, http.MethodPost, "client/register", bytes.NewBuffer(b))
	req.SetBasicAuth(cfg.UserName(), cfg.Password())

	resp := execRequest(t, cl, req)

	responseData := getData(t, expectedCode, resp)

	verifyResp(t, expectedRespData, responseData, false, func(data interface{}) []interface{} {
		var res []interface{}
		res = append(res, data.(map[string]interface{})["secret"])
		res = append(res, data.(map[string]interface{})["public_key"])
		return res
	})

	return responseData
}

func testRevokeClient(cat *clientAPITestSuite, expectedCode int, expectedRespData contract.APIResponse, clientID string) {
	revReqBody := contract.ClientRevokeRequest{ID: clientID}

	b, err := json.Marshal(&revReqBody)
	require.NoError(cat.T(), err)

	req := newRequest(cat.T(), http.MethodPost, "client/revoke", bytes.NewBuffer(b))

	ac := cat.deps.cfg.AuthConfig()
	req.SetBasicAuth(ac.UserName(), ac.Password())

	resp := execRequest(cat.T(), cat.deps.cl, req)

	responseData := getData(cat.T(), expectedCode, resp)

	verifyResp(cat.T(), expectedRespData, responseData, true, nil)
}
