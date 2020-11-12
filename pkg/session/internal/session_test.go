package internal_test

import (
	"github.com/stretchr/testify/require"
	"identification-service/pkg/session/internal"
	"testing"
)

//TODO: ADD TEST FOR BUILDER

const (
	name         = "Test Name"
	email        = "test@test.com"
	userPassword = "Password@1234"

	sessionID = "f113fe5c-de2f-4876-b734-b51fbdc96e4b"
	userID    = "86d690dd-92a0-40ac-ad48-110c951e3cb8"

	accessToken     = "v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMDozNjowNyswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTA6MjY6MDcrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiMTEwMTI0NjUtMDNhNC00OWI2LTgwODEtY2RmYzczMDlhY2MwIiwibmJmIjoiMjAyMC0xMS0wN1QxMDoyNjowNyswNTozMCJ9PrXViH5779NxXHK_PxnwW-FdFV0klU07umd8X7F0A9irFLX7GTS3AczNm_hmb_yfYOX0o4DJri89AWeCb0qTAg.bnVsbA"
	refreshToken    = "5df8159e-fd51-4e6c-9849-a9b1f070a403"
	refreshTokenTwo = "135f4d4a-48fa-4a4f-b82c-90fa8f624a16"
)

func TestCreateNewSessionSuccess(t *testing.T) {
	_, err := internal.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	require.NoError(t, err)
}

func TestCreateNewSessionFailure(t *testing.T) {
	testCases := map[string]struct {
		userID       string
		refreshToken string
	}{
		"test return error when user id is empty": {
			userID:       "",
			refreshToken: refreshToken,
		},
		"test return error when user id is invalid": {
			userID:       "invalidUserID",
			refreshToken: refreshToken,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := internal.NewSessionBuilder().UserID(testCase.userID).RefreshToken(testCase.refreshToken).Build()
			require.Error(t, err)
		})
	}
}
