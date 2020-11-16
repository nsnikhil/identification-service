package session_test

import (
	"github.com/stretchr/testify/require"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
)

func TestCreateNewSessionSuccess(t *testing.T) {
	_, err := session.NewSessionBuilder().UserID(test.UserID).RefreshToken(test.SessionRefreshToken).Build()
	require.NoError(t, err)
}

//TODO: ADD TEST CASE FOR ALL THE FAILURE SCENARIO
func TestCreateNewSessionFailure(t *testing.T) {
	testCases := map[string]struct {
		userID       string
		refreshToken string
	}{
		"test return error when user id is empty": {
			userID:       "",
			refreshToken: test.SessionRefreshToken,
		},
		"test return error when user id is invalid": {
			userID:       "invalidUserID",
			refreshToken: test.SessionRefreshToken,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := session.NewSessionBuilder().UserID(testCase.userID).RefreshToken(testCase.refreshToken).Build()
			require.Error(t, err)
		})
	}
}
