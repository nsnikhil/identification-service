package contract_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/test"
	"testing"
)

const (
	sessionEmailKey     = "email"
	sessionPasswordKey  = "password"
	sessionRefreshToken = "refreshToken"
)

var loginRequestDefaultData = map[string]string{
	sessionEmailKey:    test.RandString(8),
	sessionPasswordKey: test.RandString(8),
}

var logoutRequestDefaultData = map[string]string{
	sessionRefreshToken: test.NewUUID(),
}

var refreshTokenRequestDefaultData = map[string]string{
	sessionRefreshToken: test.NewUUID(),
}

func TestLoginRequestIsValidSuccess(t *testing.T) {
	lr := newLoginRequest(loginRequestDefaultData)
	assert.NoError(t, lr.IsValid())
}

func TestLoginRequestIsValidFailure(t *testing.T) {
	testCases := map[string]struct {
		overrides map[string]string
	}{
		"test failure when email is empty": {
			overrides: removeKey(sessionEmailKey, loginRequestDefaultData),
		},
		"test failure when old password is empty": {
			overrides: removeKey(sessionPasswordKey, loginRequestDefaultData),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			lr := newLoginRequest(testCase.overrides)
			assert.Error(t, lr.IsValid())
		})
	}
}

func TestLogoutRequestIsValidSuccess(t *testing.T) {
	lr := newLogoutRequest(logoutRequestDefaultData)
	assert.NoError(t, lr.IsValid())
}

func TestLogoutRequestIsValidFailure(t *testing.T) {
	testCases := map[string]struct {
		overrides map[string]string
	}{
		"test failure when refresh token is empty": {
			overrides: removeKey(sessionRefreshToken, logoutRequestDefaultData),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			lr := newLogoutRequest(testCase.overrides)
			assert.Error(t, lr.IsValid())
		})
	}
}

func TestRefreshTokenRequestIsValidSuccess(t *testing.T) {
	lr := newRefreshTokenRequest(refreshTokenRequestDefaultData)
	assert.NoError(t, lr.IsValid())
}

func TestRefreshTokenRequestIsValidFailure(t *testing.T) {
	testCases := map[string]struct {
		overrides map[string]string
	}{
		"test failure when refresh token is empty": {
			overrides: removeKey(sessionRefreshToken, refreshTokenRequestDefaultData),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			lr := newRefreshTokenRequest(testCase.overrides)
			assert.Error(t, lr.IsValid())
		})
	}
}

func newLoginRequest(data map[string]string) contract.LoginRequest {
	return contract.LoginRequest{
		Email:    data[sessionEmailKey],
		Password: data[sessionPasswordKey],
	}
}

func newLogoutRequest(data map[string]string) contract.LogoutRequest {
	return contract.LogoutRequest{
		RefreshToken: data[sessionRefreshToken],
	}
}

func newRefreshTokenRequest(data map[string]string) contract.RefreshTokenRequest {
	return contract.RefreshTokenRequest{
		RefreshToken: data[sessionRefreshToken],
	}
}
