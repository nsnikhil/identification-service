package user_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"testing"
)

func TestCreateNewUserSuccess(t *testing.T) {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(test.UserPasswordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, test.UserPasswordSalt).Return(test.UserPasswordKey)
	mockEncoder.On("EncodeKey", test.UserPasswordKey).Return(test.UserPasswordHash)
	mockEncoder.On("ValidatePassword", test.UserPassword).Return(nil)

	_, err := user.NewUserBuilder(mockEncoder).
		Name(test.UserName).
		Email(test.UserEmail).
		Password(test.UserPassword).
		Build()

	assert.Equal(t, nil, err)
}

//TODO: add all failure scenarios
func TestCreateNewUserValidationFailure(t *testing.T) {
	testCases := map[string]struct {
		input         func() (string, string, string)
		expectedError error
	}{
		"test failure when name is empty": {
			input: func() (string, string, string) {
				return test.EmptyString, test.UserEmail, test.UserPassword
			},
			expectedError: liberr.WithArgs(errors.New("name cannot be empty")),
		},

		"test failure when email is empty": {
			input: func() (string, string, string) {
				return test.UserEmail, test.EmptyString, test.UserPassword
			},
			expectedError: liberr.WithArgs(errors.New("email cannot be empty")),
		},

		"test failure when password is empty": {
			input: func() (string, string, string) {
				return test.UserEmail, test.UserEmail, test.EmptyString
			},
			expectedError: liberr.WithArgs(errors.New("password cannot be empty")),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			name, email, userPassword := testCase.input()
			_, err := user.NewUserBuilder(&password.MockEncoder{}).
				Name(name).
				Email(email).
				Password(userPassword).
				Build()

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

	_, err := user.NewUserBuilder(mockEncoder).
		Name(test.UserName).
		Email(test.UserEmail).
		Password(test.UserPasswordInvalid).
		Build()

	assert.Error(t, err)
}
