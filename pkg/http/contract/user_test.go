package contract_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/test"
	"testing"
)

const (
	userEmailKey       = "email"
	userOldPasswordKey = "oldPassword"
	userNewPasswordKey = "newPassword"
)

var updatePasswordDefaultData = map[string]string{
	userEmailKey:       test.RandString(8),
	userOldPasswordKey: test.RandString(8),
	userNewPasswordKey: test.RandString(8),
}

func TestUpdatePasswordRequestIsValidSuccess(t *testing.T) {
	upr := newUpdatePasswordRequest(updatePasswordDefaultData)
	assert.NoError(t, upr.IsValid())
}

func TestUpdatePasswordRequestIsValidFailure(t *testing.T) {
	testCases := map[string]struct {
		overrides map[string]string
	}{
		"test failure when email is empty": {
			overrides: removeKey(userEmailKey, updatePasswordDefaultData),
		},
		"test failure when old password is empty": {
			overrides: removeKey(userOldPasswordKey, updatePasswordDefaultData),
		},
		"test failure when new password is empty": {
			overrides: removeKey(userNewPasswordKey, updatePasswordDefaultData),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			upr := newUpdatePasswordRequest(testCase.overrides)
			assert.Error(t, upr.IsValid())
		})
	}
}

func newUpdatePasswordRequest(data map[string]string) contract.UpdatePasswordRequest {
	return contract.UpdatePasswordRequest{
		Email:       data[userEmailKey],
		OldPassword: data[userOldPasswordKey],
		NewPassword: data[userNewPasswordKey],
	}
}
