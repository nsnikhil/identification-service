package contract_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/test"
	"testing"
)

const (
	emailKey       = "email"
	oldPasswordKey = "oldPassword"
	newPasswordKey = "newPassword"
)

var defaultData = map[string]string{
	emailKey:       test.RandString(8),
	oldPasswordKey: test.RandString(8),
	newPasswordKey: test.RandString(8),
}

func TestUpdatePasswordRequestIsValidSuccess(t *testing.T) {
	upr := newUpdatePasswordRequest(defaultData)
	assert.NoError(t, upr.IsValid())
}

func TestUpdatePasswordRequestIsValidFailure(t *testing.T) {
	testCases := map[string]struct {
		overrides map[string]string
	}{
		"test failure when email is empty": {
			overrides: removeKey(emailKey, defaultData),
		},
		"test failure when old password is empty": {
			overrides: removeKey(oldPasswordKey, defaultData),
		},
		"test failure when new password is empty": {
			overrides: removeKey(newPasswordKey, defaultData),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			upr := newUpdatePasswordRequest(testCase.overrides)
			assert.Error(t, upr.IsValid())
		})
	}
}

func removeKey(key string, defaultData map[string]string) map[string]string {
	temp := make(map[string]string)

	for k, v := range defaultData {
		if k != key {
			temp[k] = v
		}
	}

	return temp
}

func newUpdatePasswordRequest(data map[string]string) contract.UpdatePasswordRequest {
	return contract.UpdatePasswordRequest{
		Email:       data[emailKey],
		OldPassword: data[oldPasswordKey],
		NewPassword: data[newPasswordKey],
	}
}
