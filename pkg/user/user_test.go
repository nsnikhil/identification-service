package user_test

import (
	"errors"
	"github.com/nsnikhil/erx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/password"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"testing"
	"time"
)

const (
	idKey               = "id"
	nameKey             = "name"
	emailKey            = "email"
	userPasswordKey     = "userPassword"
	userPasswordSaltKey = "userPasswordSalt"
	userPasswordHashKey = "userPasswordHash"
	createdAtKey        = "createdAt"
	updatedAtKey        = "updatedAt"
)

func TestCreateNewUserSuccess(t *testing.T) {
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	_, err := user.NewUserBuilder(mockEncoder).
		Name(test.RandString(8)).
		Email(test.NewEmail()).
		Password(userPassword).
		Build()

	assert.Equal(t, nil, err)
}

func TestCreateNewUserValidationFailure(t *testing.T) {
	testCases := map[string]map[string]interface{}{
		"test failure when id is empty":                     {idKey: ""},
		"test failure when id is invalid":                   {idKey: "invalid id"},
		"test failure when name is empty":                   {nameKey: ""},
		"test failure when email is empty":                  {emailKey: ""},
		"test failure when password is empty":               {userPasswordKey: ""},
		"test failure when password hash is empty":          {userPasswordHashKey: ""},
		"test failure when password salt is empty":          {userPasswordSaltKey: []byte{}},
		"test failure when created at is set to zero value": {createdAtKey: time.Time{}},
		"test failure when updated at is set to zero value": {updatedAtKey: time.Time{}},
	}

	for name, data := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := buildUser(data)
			assert.Error(t, err)
		})
	}
}

func buildUser(d map[string]interface{}) (user.User, error) {
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	either := func(a interface{}, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return user.NewUserBuilder(mockEncoder).
		ID(either(d[idKey], test.NewUUID()).(string)).
		Name(either(d[nameKey], test.RandString(8)).(string)).
		Email(either(d[emailKey], test.NewEmail()).(string)).
		Password(either(d[userPasswordKey], userPassword).(string)).
		PasswordSalt(either(d[userPasswordSaltKey], test.RandBytes(86)).([]byte)).
		PasswordHash(either(d[userPasswordHashKey], test.RandString(44)).(string)).
		CreatedAt(either(d[createdAtKey], test.CreatedAt).(time.Time)).
		UpdatedAt(either(d[updatedAtKey], test.UpdatedAt).(time.Time)).
		Build()
}

func TestCreateNewUserFailureForInvalidPassword(t *testing.T) {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"ValidatePassword",
		mock.AnythingOfType("string"),
	).Return(erx.WithArgs(errors.New("invalid password")))

	_, err := user.NewUserBuilder(mockEncoder).
		Name(test.RandString(8)).
		Email(test.NewEmail()).
		Password(test.RandString(12)).
		Build()

	assert.Error(t, err)
}
