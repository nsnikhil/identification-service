// build component_test

package test

import (
	"bytes"
	"encoding/json"
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

func (cat *clientAPITestSuite) AfterTest(suiteName, testName string) {
	truncateTables(cat.T(), cat.deps.ctx, cat.deps.db, ClientTableName)
}

func (cat *clientAPITestSuite) TearDownSuite() {
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

//TODO: WHY DOES DUPLICATE RECORD RETURN INTERNAL SERVER ERROR ?
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
		http.StatusInternalServerError,
		contract.APIResponse{
			Success: false, Error: &contract.Error{Message: "internal server error"},
		},
		reqBody,
	)
}

func (cat *clientAPITestSuite) TestRevokeClientSuccess() {
	registerClientAndGetHeaders(cat.T(), cat.deps.cfg.AuthConfig(), cat.deps.cl)

	var clientID string

	err := cat.deps.db.QueryRowContext(
		cat.deps.ctx, `select id from clients where name = $1`, ClientName,
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

	testRevokeClient(cat, http.StatusNotFound, expectedRespData, ClientID)
}

func getRegisterClientReqBody(data map[string]interface{}) contract.CreateClientRequest {
	either := func(a, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return contract.CreateClientRequest{
		Name:              either(data[clientNameKey], ClientName).(string),
		AccessTokenTTL:    either(data[clientAccessTokenTTLKey], ClientAccessTokenTTL).(int),
		SessionTTL:        either(data[clientSessionTTLKey], ClientSessionTTL).(int),
		MaxActiveSessions: either(data[clientMaxActiveSessionsKey], ClientMaxActiveSessions).(int),
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
