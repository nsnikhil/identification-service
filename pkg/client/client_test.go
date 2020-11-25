package client_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/test"
	"testing"
	"time"
)

const (
	id                  = "id"
	name                = "name"
	secret              = "secret"
	revoked             = "revoked"
	accessTokenTTL      = "accessTokenTTL"
	sessionTTL          = "sessionTTL"
	maxActiveSessions   = "maxActiveSessions"
	sessionStrategyName = "sessionStrategyName"
	privateKey          = "privateKey"
	createdAt           = "createdAt"
	updatedAt           = "updatedAt"
)

func TestClientBuilderBuildSuccess(t *testing.T) {
	test.NewClient(t)
}

func TestClientBuilderBuildFailureValidation(t *testing.T) {
	testCases := map[string]map[string]interface{}{
		"test failure when id is empty":                        {id: ""},
		"test failure when id is invalid":                      {id: "invalid id"},
		"test failure when name is empty":                      {name: ""},
		"test failure when secret is empty":                    {secret: ""},
		"test failure when secret is invalid":                  {secret: "invalid secret"},
		"test failure when access token ttl is less than 1":    {accessTokenTTL: 0},
		"test failure when session ttl is less than 1":         {sessionTTL: 0},
		"test failure when max active sessions is less than 1": {maxActiveSessions: 0},
		"test failure when session strategy is empty":          {sessionStrategyName: ""},
		"test failure when private key is empty":               {privateKey: []byte{}},
		"test failure when created at is set to zero value":    {createdAt: time.Time{}},
		"test failure when updated at is set to zero value":    {updatedAt: time.Time{}},
	}

	for name, data := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := buildClient(data)
			assert.Error(t, err)
		})
	}
}

func buildClient(d map[string]interface{}) (client.Client, error) {
	either := func(a interface{}, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return client.NewClientBuilder().
		ID(either(d[id], test.ClientID).(string)).
		Name(either(d[name], test.ClientName).(string)).
		Secret(either(d[secret], test.ClientSecret).(string)).
		Revoked(either(d[revoked], false).(bool)).
		AccessTokenTTL(either(d[accessTokenTTL], test.ClientAccessTokenTTL).(int)).
		SessionTTL(either(d[sessionTTL], test.ClientSessionTTL).(int)).
		MaxActiveSessions(either(d[maxActiveSessions], test.ClientMaxActiveSessions).(int)).
		SessionStrategy(either(d[sessionStrategyName], test.ClientSessionStrategyRevokeOld).(string)).
		PrivateKey(either(d[privateKey], test.ClientPriKeyBytes).([]byte)).
		CreatedAt(either(d[createdAt], test.CreatedAt).(time.Time)).
		UpdatedAt(either(d[updatedAt], test.UpdatedAt).(time.Time)).
		Build()
}

func TestClientWithContextSuccess(t *testing.T) {
	_, err := client.WithContext(context.Background(), test.NewClient(t))
	assert.Nil(t, err)
}

func TestClientWithContextFailure(t *testing.T) {
	_, err := client.WithContext(context.Background(), client.Client{})
	assert.Error(t, err)
}

func TestClientFromContextSuccess(t *testing.T) {
	ctx, err := client.WithContext(context.Background(), test.NewClient(t))
	assert.Nil(t, err)

	_, err = client.FromContext(ctx)
	assert.Nil(t, err)
}

func TestClientFromContextFailure(t *testing.T) {
	_, err := client.FromContext(context.Background())
	assert.Error(t, err)
}

func TestClientGetters(t *testing.T) {
	cl, err := buildClient(map[string]interface{}{})
	require.NoError(t, err)

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
			expectedData: test.ClientAccessTokenTTL,
		},
		"test get session ttl": {
			actualData:   cl.SessionTTL(),
			expectedData: test.ClientSessionTTL,
		},
		"test get session strategy name": {
			actualData:   cl.SessionStrategyName(),
			expectedData: test.ClientSessionStrategyRevokeOld,
		},
		"test get max active sessions": {
			actualData:   cl.MaxActiveSessions(),
			expectedData: test.ClientMaxActiveSessions,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedData, testCase.actualData)
		})
	}
}
