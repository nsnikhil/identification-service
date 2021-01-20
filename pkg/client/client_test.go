package client_test

import (
	"context"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/test"
	"testing"
	"time"
)

type clientTest struct {
	suite.Suite
	cfg         config.ClientConfig
	defaultData map[string]interface{}
}

func (ct *clientTest) SetupSuite() {
	mockClientConfig := &config.MockClientConfig{}
	mockClientConfig.On("Strategies").
		Return(map[string]bool{test.ClientSessionStrategyRevokeOld: true})

	ct.cfg = mockClientConfig
	ct.defaultData = map[string]interface{}{}
}

func TestClient(t *testing.T) {
	suite.Run(t, new(clientTest))
}

func (ct *clientTest) TestClientBuilderBuildSuccess() {
	_, err := test.NewClient(ct.cfg, ct.defaultData)
	ct.Require().NoError(err)
}

func (ct *clientTest) TestClientBuilderBuildFailureValidation() {
	testCases := map[string]map[string]interface{}{
		"test failure when id is empty":                        {test.ClientIdKey: ""},
		"test failure when id is invalid":                      {test.ClientIdKey: "invalid id"},
		"test failure when name is empty":                      {test.ClientNameKey: ""},
		"test failure when secret is empty":                    {test.ClientSecretKey: ""},
		"test failure when secret is invalid":                  {test.ClientSecretKey: "invalid secret"},
		"test failure when access token ttl is less than 1":    {test.ClientAccessTokenTTLKey: 0},
		"test failure when session ttl is less than 1":         {test.ClientSessionTTLKey: 0},
		"test failure when max active sessions is less than 1": {test.ClientMaxActiveSessionsKey: 0},
		"test failure when session strategy is empty":          {test.ClientSessionStrategyNameKey: ""},
		"test failure when session strategy is invalid":        {test.ClientSessionStrategyNameKey: "invalid"},
		"test failure when private key is empty":               {test.ClientPrivateKeyKey: []byte{}},
		"test failure when created at is set to zero value":    {test.ClientCreatedAtKey: time.Time{}},
		"test failure when updated at is set to zero value":    {test.ClientUpdatedAtKey: time.Time{}},
	}

	for name, data := range testCases {
		ct.Run(name, func() {
			_, err := test.NewClient(ct.cfg, data)
			ct.Require().Error(err)
		})
	}
}

func (ct *clientTest) TestClientWithContextSuccess() {
	cl, err := test.NewClient(ct.cfg, ct.defaultData)
	ct.Require().NoError(err)

	_, err = client.WithContext(context.Background(), cl)
	ct.Assert().NoError(err)
}

func (ct *clientTest) TestClientWithContextFailure() {
	_, err := client.WithContext(context.Background(), client.Client{})
	ct.Assert().Error(err)
}

func (ct *clientTest) TestClientFromContextSuccess() {
	cl, err := test.NewClient(ct.cfg, ct.defaultData)
	ct.Require().NoError(err)

	ctx, err := client.WithContext(context.Background(), cl)
	ct.Require().NoError(err)

	_, err = client.FromContext(ctx)
	ct.Require().NoError(err)
}

func (ct *clientTest) TestClientFromContextFailure() {
	_, err := client.FromContext(context.Background())
	ct.Assert().Error(err)
}

func (ct *clientTest) TestClientGetters() {
	accessTokenTTLVal := test.RandInt(1, 10)
	sessionTTLVal := test.RandInt(1440, 86701)
	maxActiveSessionsVal := test.RandInt(1, 10)

	cl, err := test.NewClient(
		ct.cfg,
		map[string]interface{}{
			test.ClientAccessTokenTTLKey:    accessTokenTTLVal,
			test.ClientSessionTTLKey:        sessionTTLVal,
			test.ClientMaxActiveSessionsKey: maxActiveSessionsVal,
		},
	)
	ct.Require().NoError(err)

	testCases := map[string]struct {
		actualData   interface{}
		expectedData interface{}
	}{
		"test get revoked": {
			actualData:   cl.IsRevoked(),
			expectedData: false,
		},
		"test get access token ttl": {
			actualData:   cl.AccessTokenTTL(),
			expectedData: accessTokenTTLVal,
		},
		"test get session ttl": {
			actualData:   cl.SessionTTL(),
			expectedData: sessionTTLVal,
		},
		"test get session strategy name": {
			actualData:   cl.SessionStrategyName(),
			expectedData: test.ClientSessionStrategyRevokeOld,
		},
		"test get max active sessions": {
			actualData:   cl.MaxActiveSessions(),
			expectedData: maxActiveSessionsVal,
		},
	}

	for name, testCase := range testCases {
		ct.Run(name, func() {
			ct.Assert().Equal(testCase.expectedData, testCase.actualData)
		})
	}
}
