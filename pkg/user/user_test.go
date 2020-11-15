package user_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/user"
	"testing"
)

const (
	name               = "Test Name"
	email              = "test@test.com"
	userPassword       = "Password@1234"
	newPassword        = "NewPassword@1234"
	invalidPasswordOne = "password@1234"

	emptyString = ""

	invalidPassword = "password@1234"
	userID          = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
)

var salt = []byte{90, 20, 247, 194, 220, 48, 153, 58, 158, 103, 9, 17, 243, 24, 179, 254, 88, 59, 161, 81, 216, 8, 126, 122, 102, 151, 200, 12, 134, 118, 146, 197, 193, 248, 117, 57, 127, 137, 112, 233, 116, 50, 128, 84, 127, 93, 180, 23, 81, 69, 245, 183, 45, 57, 51, 125, 9, 46, 200, 175, 97, 49, 11, 0, 40, 228, 186, 60, 177, 43, 69, 52, 168, 195, 69, 101, 21, 245, 62, 131, 252, 96, 240, 154, 251, 2}
var key = []byte{34, 179, 107, 154, 0, 94, 48, 1, 134, 44, 128, 127, 254, 17, 124, 248, 69, 96, 196, 174, 146, 255, 131, 91, 94, 143, 105, 33, 230, 157, 77, 243}
var hash = "IrNrmgBeMAGGLIB//hF8+EVgxK6S/4NbXo9pIeadTfM="

func TestCreateNewUserSuccess(t *testing.T) {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(salt, nil)
	mockEncoder.On("GenerateKey", userPassword, salt).Return(key)
	mockEncoder.On("EncodeKey", key).Return(hash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	_, err := user.NewUser(mockEncoder, name, email, userPassword)
	assert.Equal(t, nil, err)
}

func TestCreateNewUserValidationFailure(t *testing.T) {
	testCases := map[string]struct {
		input         func() (string, string, string)
		expectedError error
	}{
		"test failure when name is empty": {
			input: func() (string, string, string) {
				return emptyString, email, userPassword
			},
			expectedError: liberr.WithArgs(errors.New("name cannot be empty")),
		},

		"test failure when email is empty": {
			input: func() (string, string, string) {
				return name, emptyString, userPassword
			},
			expectedError: liberr.WithArgs(errors.New("email cannot be empty")),
		},

		"test failure when password is empty": {
			input: func() (string, string, string) {
				return name, email, emptyString
			},
			expectedError: liberr.WithArgs(errors.New("password cannot be empty")),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			name, email, userPassword := testCase.input()
			_, err := user.NewUser(&password.MockEncoder{}, name, email, userPassword)
			assert.Error(t, err)
		})
	}
}

func TestCreateNewUserFailureForInvalidPassword(t *testing.T) {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"ValidatePassword",
		mock.AnythingOfType("string"),
	).Return(liberr.WithArgs(errors.New("invalid password")))

	_, err := user.NewUser(mockEncoder, name, email, invalidPassword)
	assert.Error(t, err)
}
