package client_test

import (
	"crypto/ed25519"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"testing"
)

const (
	name              = "clientOne"
	accessTokenTTL    = 10
	sessionTTL        = 14400
	maxActiveSessions = 2
	id                = "f14abb31-ec1a-4ff6-a937-c2e930ca34ef"
	secret            = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
)

var pubKey = ed25519.PublicKey{6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}
var priKey = ed25519.PrivateKey{3, 195, 208, 247, 190, 104, 63, 62, 164, 50, 63, 217, 229, 215, 179, 62, 223, 104, 197, 43, 164, 164, 231, 1, 22, 70, 154, 130, 109, 98, 88, 210, 6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}

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
			accessTokenTTL: accessTokenTTL,
			sessionTTL:     sessionTTL,
		},
		"test failure when access token ttl is below 1": {
			name:           name,
			accessTokenTTL: 0,
			sessionTTL:     sessionTTL,
		},
		"test failure when session ttl is below 1": {
			name:           name,
			accessTokenTTL: accessTokenTTL,
			sessionTTL:     0,
		},
		"test failure when session ttl is less than access token ttl": {
			name:           name,
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
		Name(name).
		AccessTokenTTL(accessTokenTTL).
		SessionTTL(sessionTTL).
		MaxActiveSessions(maxActiveSessions).
		PrivateKey(priKey).
		Build()

	require.NoError(t, err)

	return cl
}
