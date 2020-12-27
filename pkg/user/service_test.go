package user_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/event/publisher"
	"identification-service/pkg/password"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"testing"
)

type createUserSuite struct {
	encoder   password.Encoder
	publisher publisher.Publisher
	suite.Suite
}

func (cst *createUserSuite) SetupSuite() {
	passwordSalt := test.UserPasswordSalt()
	passwordKey := test.UserPasswordKey()
	passwordHash := test.UserPasswordHash()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)

	mockPublisher := &publisher.MockPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	cst.encoder = mockEncoder
	cst.publisher = mockPublisher
}

func (cst *createUserSuite) TestCreateUserSuccess() {
	mockStore := &user.MockStore{}
	mockStore.On(
		"CreateUser",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("User"),
	).Return(test.UserID(), nil)

	cst.encoder.(*password.MockEncoder).On("ValidatePassword", test.UserPassword).Return(nil)

	service := user.NewService(mockStore, cst.encoder, cst.publisher)

	_, err := service.CreateUser(context.Background(), test.UserName(), test.UserEmail(), test.UserPassword)
	assert.Nil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenStoreCallFails() {
	mockStore := &user.MockStore{}
	mockStore.On(
		"CreateUser",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("User"),
	).Return("", errors.New("failed to save new user"))

	cst.encoder.(*password.MockEncoder).On("ValidatePassword", test.UserPassword).Return(nil)

	service := user.NewService(mockStore, cst.encoder, cst.publisher)

	_, err := service.CreateUser(context.Background(), test.UserName(), test.UserEmail(), test.UserPassword)
	assert.NotNil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenInputIsInvalid() {
	testCases := map[string]struct {
		input func() (string, string, string)
		err   error
	}{
		"test failure when name is empty": {
			input: func() (string, string, string) {
				return test.EmptyString, test.UserEmail(), test.UserPassword
			},
			err: errors.New("name cannot be empty"),
		},
		"test failure when email is empty": {
			input: func() (string, string, string) {
				return test.UserName(), test.EmptyString, test.UserPassword
			},
			err: errors.New("email cannot be empty"),
		},
		"test failure when pass is empty": {
			input: func() (string, string, string) {
				return test.UserName(), test.UserEmail(), test.EmptyString
			},
			err: errors.New("password cannot be empty"),
		},
		"test failure when password is invalid": {
			input: func() (string, string, string) {
				cst.encoder.(*password.MockEncoder).On(
					"ValidatePassword",
					test.UserPasswordInvalid,
				).Return(errors.New("invalid password"))
				return test.UserName(), test.UserEmail(), test.UserPasswordInvalid
			},
			err: errors.New("invalid password"),
		},
	}

	for name, testCase := range testCases {
		cst.T().Run(name, func(t *testing.T) {
			service := user.NewService(&user.MockStore{}, cst.encoder, &publisher.MockPublisher{})

			name, email, userPassword := testCase.input()
			_, err := service.CreateUser(context.Background(), name, email, userPassword)
			assert.NotNil(cst.T(), err)
		})
	}

}

func TestCreateUser(t *testing.T) {
	suite.Run(t, new(createUserSuite))
}

func TestGetUserIDSuccess(t *testing.T) {
	userEmail := test.UserEmail()

	mockStore := &user.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.emptyCtx"), userEmail).Return(user.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"VerifyPassword",
		test.UserPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil)

	service := user.NewService(mockStore, mockEncoder, &publisher.MockPublisher{})

	_, err := service.GetUserID(context.Background(), userEmail, test.UserPassword)
	require.NoError(t, err)
}

func TestGetUserIDFailureWhenStoreCallsFails(t *testing.T) {
	userEmail := test.UserEmail()

	mockStore := &user.MockStore{}
	mockStore.On(
		"GetUser",
		mock.AnythingOfType("*context.emptyCtx"),
		userEmail,
	).Return(user.User{}, errors.New("failed to get user"))

	service := user.NewService(mockStore, &password.MockEncoder{}, &publisher.MockPublisher{})

	_, err := service.GetUserID(context.Background(), userEmail, test.UserPassword)
	require.Error(t, err)
}

func TestGetUserIDFailureWhenPasswordVerificationFails(t *testing.T) {
	userEmail := test.UserEmail()

	mockStore := &user.MockStore{}
	mockStore.On(
		"GetUser",
		mock.AnythingOfType("*context.emptyCtx"),
		userEmail,
	).Return(user.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"VerifyPassword",
		test.UserPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(errors.New("invalid credentials"))

	service := user.NewService(mockStore, mockEncoder, &publisher.MockPublisher{})

	_, err := service.GetUserID(context.Background(), userEmail, test.UserPassword)
	require.Error(t, err)
}

func TestUpdatePasswordSuccess(t *testing.T) {
	userEmail := test.UserEmail()
	passwordSalt := test.UserPasswordSalt()
	passwordKey := test.UserPasswordKey()
	passwordHash := test.UserPasswordHash()

	mockStore := &user.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.emptyCtx"), userEmail).Return(user.User{}, nil)
	mockStore.On(
		"UpdatePassword",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"), passwordHash, passwordSalt,
	).Return(int64(1), nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPasswordNew, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("VerifyPassword",
		test.UserPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil)
	mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)

	mockPublisher := &publisher.MockPublisher{}
	mockPublisher.On("Publish", mock.Anything, mock.Anything).Return(nil)

	service := user.NewService(mockStore, mockEncoder, mockPublisher)

	err := service.UpdatePassword(context.Background(), userEmail, test.UserPassword, test.UserPasswordNew)
	require.NoError(t, err)
}

func TestUpdatePasswordFailure(t *testing.T) {
	userEmail := test.UserEmail()
	passwordSalt := test.UserPasswordSalt()
	passwordKey := test.UserPasswordKey()
	passwordHash := test.UserPasswordHash()

	testCases := map[string]struct {
		store       func() user.Store
		encoder     func() password.Encoder
		newPassword string
	}{
		"test failure when new password does not match spec": {
			store: func() user.Store { return &user.MockStore{} },
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On(
					"ValidatePassword",
					test.UserPasswordInvalid,
				).Return(errors.New("invalid password"))

				return mockEncoder
			},
			newPassword: test.UserPasswordInvalid,
		},
		"test failure when get user fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.emptyCtx"),
					userEmail,
				).Return(user.User{}, errors.New("failed to get user"))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)

				return mockEncoder
			},
			newPassword: test.UserPasswordNew,
		},
		"test failure when generate salt fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.emptyCtx"),
					userEmail,
				).Return(user.User{}, nil)

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)
				mockEncoder.On(
					"VerifyPassword",
					test.UserPassword,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				mockEncoder.On("GenerateSalt").Return(
					passwordSalt,
					errors.New("failed to generate salt"),
				)

				return mockEncoder
			},
			newPassword: test.UserPasswordNew,
		},
		"test failure when generate store call fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.emptyCtx"),
					userEmail,
				).Return(user.User{}, nil)
				mockStore.On(
					"UpdatePassword",
					mock.AnythingOfType("*context.emptyCtx"),
					mock.AnythingOfType("string"),
					passwordHash,
					passwordSalt,
				).Return(int64(0), errors.New("failed to update password"))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)
				mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
				mockEncoder.On("GenerateKey", test.UserPasswordNew, passwordSalt).Return(passwordKey)
				mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
				mockEncoder.On(
					"VerifyPassword",
					test.UserPassword,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				return mockEncoder
			},
			newPassword: test.UserPasswordNew,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := user.NewService(testCase.store(), testCase.encoder(), &publisher.MockPublisher{})

			err := service.UpdatePassword(context.Background(), userEmail, test.UserPassword, testCase.newPassword)
			require.Error(t, err)
		})
	}
}
