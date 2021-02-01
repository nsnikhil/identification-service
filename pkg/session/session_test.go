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
	idKey           = "id"
	userIDKey       = "userID"
	refreshTokenKey = "refreshToken"
	revokedKey      = "revoked"
	createdAtKey    = "createdAt"
	updatedAtKey    = "updatedAt"
)

func TestCreateNewSessionSuccess(t *testing.T) {
	_, err := session.NewSessionBuilder().UserID(test.NewUUID()).RefreshToken(test.NewUUID()).Build()
	require.NoError(t, err)
}

func TestCreateNewSessionValidationFailure(t *testing.T) {
	testCases := map[string]map[string]interface{}{
		"test failure when id is empty":                     {idKey: ""},
		"test failure when id is invalid":                   {idKey: "invalid id"},
		"test failure when userID is empty":                 {userIDKey: ""},
		"test failure when userID is invalid":               {userIDKey: "invalid id"},
		"test failure when refreshToken is empty":           {refreshTokenKey: ""},
		"test failure when refreshToken is invalid":         {refreshTokenKey: "invalid id"},
		"test failure when created at is set to zero value": {createdAtKey: time.Time{}},
		"test failure when updated at is set to zero value": {updatedAtKey: time.Time{}},
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
		ID(either(d[idKey], test.NewUUID()).(string)).
		UserID(either(d[userIDKey], test.NewUUID()).(string)).
		RefreshToken(either(d[refreshTokenKey], test.NewUUID()).(string)).
		Revoked(either(d[revokedKey], false).(bool)).
		CreatedAt(either(d[createdAtKey], test.CreatedAt).(time.Time)).
		UpdatedAt(either(d[updatedAtKey], test.UpdatedAt).(time.Time)).
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
		CreatedAt(time.Now().AddDate(0, -2, 2)).
		Build()

	assert.Nil(t, err)

	assert.False(t, ss.IsExpired(87600))
}
