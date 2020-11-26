package session_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
	"time"
)

const (
	id           = "id"
	userID       = "userID"
	refreshToken = "refreshToken"
	revoked      = "revoked"
	createdAt    = "createdAt"
	updatedAt    = "updatedAt"
)

func TestCreateNewSessionSuccess(t *testing.T) {
	_, err := session.NewSessionBuilder().UserID(test.UserID).RefreshToken(test.SessionRefreshToken).Build()
	require.NoError(t, err)
}

func TestCreateNewSessionValidationFailure(t *testing.T) {
	testCases := map[string]map[string]interface{}{
		"test failure when id is empty":                     {id: ""},
		"test failure when id is invalid":                   {id: "invalid id"},
		"test failure when userID is empty":                 {userID: ""},
		"test failure when userID is invalid":               {userID: "invalid id"},
		"test failure when refreshToken is empty":           {refreshToken: ""},
		"test failure when refreshToken is invalid":         {refreshToken: "invalid id"},
		"test failure when created at is set to zero value": {createdAt: time.Time{}},
		"test failure when updated at is set to zero value": {updatedAt: time.Time{}},
	}

	for name, data := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := buildSession(data)
			assert.Error(t, err)
		})
	}
}

func buildSession(d map[string]interface{}) (session.Session, error) {
	either := func(a interface{}, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return session.NewSessionBuilder().
		ID(either(d[id], test.SessionID).(string)).
		UserID(either(d[userID], test.UserID).(string)).
		RefreshToken(either(d[refreshToken], test.SessionRefreshToken).(string)).
		Revoked(either(d[revoked], false).(bool)).
		CreatedAt(either(d[createdAt], test.CreatedAt).(time.Time)).
		UpdatedAt(either(d[updatedAt], test.UpdatedAt).(time.Time)).
		Build()
}

func TestIsExpiredTrue(t *testing.T) {
	ss, err := session.NewSessionBuilder().
		CreatedAt(time.Now().AddDate(0, -2, -1)).
		Build()

	assert.Nil(t, err)

	assert.True(t, ss.IsExpired(87600))
}

func TestIsExpiredFalse(t *testing.T) {
	ss, err := session.NewSessionBuilder().
		CreatedAt(time.Now().AddDate(0, -2, 1)).
		Build()

	assert.Nil(t, err)

	assert.False(t, ss.IsExpired(87600))
}
