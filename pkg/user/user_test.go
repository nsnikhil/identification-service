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
	"time"
)

const (
	id               = "id"
	name             = "name"
	email            = "email"
	userPassword     = "userPassword"
	userPasswordSalt = "userPasswordSalt"
	userPasswordHash = "userPasswordHash"
	createdAt        = "createdAt"
	updatedAt        = "updatedAt"
)

func TestCreateNewUserSuccess(t *testing.T) {
	passwordSalt := test.UserPasswordSalt()
	passwordKey := test.UserPasswordKey()
	passwordHash := test.UserPasswordHash()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", test.UserPassword).Return(nil)

	_, err := user.NewUserBuilder(mockEncoder).
		Name(test.UserName()).
		Email(test.UserEmail()).
		Password(test.UserPassword).
		Build()

	assert.Equal(t, nil, err)
}

func TestCreateNewUserValidationFailure(t *testing.T) {
	testCases := map[string]map[string]interface{}{
		"test failure when id is empty":                     {id: ""},
		"test failure when id is invalid":                   {id: "invalid id"},
		"test failure when name is empty":                   {name: ""},
		"test failure when email is empty":                  {email: ""},
		"test failure when password is empty":               {userPassword: ""},
		"test failure when password hash is empty":          {userPasswordHash: ""},
		"test failure when password salt is empty":          {userPasswordSalt: []byte{}},
		"test failure when created at is set to zero value": {createdAt: time.Time{}},
		"test failure when updated at is set to zero value": {updatedAt: time.Time{}},
	}

	for name, data := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := buildUser(data)
			assert.Error(t, err)
		})
	}
}

func buildUser(d map[string]interface{}) (user.User, error) {
	passwordSalt := test.UserPasswordSalt()
	passwordKey := test.UserPasswordKey()
	passwordHash := test.UserPasswordHash()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", test.UserPassword).Return(nil)

	either := func(a interface{}, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return user.NewUserBuilder(mockEncoder).
		ID(either(d[id], test.UserID()).(string)).
		Name(either(d[name], test.UserName()).(string)).
		Email(either(d[email], test.UserEmail()).(string)).
		Password(either(d[userPassword], test.UserPassword).(string)).
		PasswordSalt(either(d[userPasswordSalt], test.UserPasswordSalt()).([]byte)).
		PasswordHash(either(d[userPasswordHash], test.UserPasswordHash()).(string)).
		CreatedAt(either(d[createdAt], test.CreatedAt).(time.Time)).
		UpdatedAt(either(d[updatedAt], test.UpdatedAt).(time.Time)).
		Build()
}

func TestCreateNewUserFailureForInvalidPassword(t *testing.T) {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"ValidatePassword",
		mock.AnythingOfType("string"),
	).Return(liberr.WithArgs(errors.New("invalid password")))

	_, err := user.NewUserBuilder(mockEncoder).
		Name(test.UserName()).
		Email(test.UserEmail()).
		Password(test.UserPasswordInvalid).
		Build()

	assert.Error(t, err)
}
