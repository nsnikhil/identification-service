package client_test

import (
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/test"
	"testing"
)

func TestClientBuilderBuildSuccess(t *testing.T) {
	createNewClient(t)
}

//TODO: ADD ALL TESTCASES FOR CLIENT
func TestClientBuilderBuildFailure(t *testing.T) {
	testCases := map[string]struct {
		name           string
		accessTokenTTL int
		sessionTTL     int
	}{
		"test failure when name is empty": {
			name:           "",
			accessTokenTTL: test.ClientAccessTokenTTL,
			sessionTTL:     test.ClientSessionTTL,
		},
		"test failure when access token ttl is below 1": {
			name:           test.ClientName,
			accessTokenTTL: 0,
			sessionTTL:     test.ClientSessionTTL,
		},
		"test failure when session ttl is below 1": {
			name:           test.ClientName,
			accessTokenTTL: test.ClientAccessTokenTTL,
			sessionTTL:     0,
		},
		"test failure when session ttl is less than access token ttl": {
			name:           test.ClientName,
			accessTokenTTL: 20,
			sessionTTL:     10,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := client.NewClientBuilder().
				Name(testCase.name).
				AccessTokenTTL(testCase.accessTokenTTL).
				SessionTTL(testCase.sessionTTL).
				Build()

			require.Error(t, err)
		})
	}
}

func createNewClient(t *testing.T) client.Client {
	cl, err := client.NewClientBuilder().
		Name(test.ClientName).
		AccessTokenTTL(test.ClientAccessTokenTTL).
		SessionTTL(test.ClientSessionTTL).
		MaxActiveSessions(test.ClientMaxActiveSessions).
		PrivateKey(test.ClientPriKey).
		Build()

	require.NoError(t, err)

	return cl
}
