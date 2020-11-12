package internal_test

import (
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client/internal"
	"testing"
)

const (
	name           = "clientOne"
	accessTokenTTL = 10
	sessionTTL     = 14400
)

func TestCreateNewClientSuccess(t *testing.T) {
	_, err := internal.NewClientBuilder().Name(name).AccessTokenTTL(accessTokenTTL).SessionTTL(sessionTTL).Build()
	require.NoError(t, err)
}

func TestCreateNewClientFailure(t *testing.T) {
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
			_, err := internal.NewClientBuilder().Name(testCase.name).AccessTokenTTL(testCase.accessTokenTTL).SessionTTL(testCase.sessionTTL).Build()
			require.Error(t, err)
		})
	}
}
