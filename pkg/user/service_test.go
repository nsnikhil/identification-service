package user_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/queue"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"testing"
)

type createUserSuite struct {
	encoder password.Encoder
	queue   queue.Queue
	suite.Suite
}

func (cst *createUserSuite) SetupSuite() {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(test.UserPasswordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPassword, test.UserPasswordSalt).Return(test.UserPasswordKey)
	mockEncoder.On("EncodeKey", test.UserPasswordKey).Return(test.UserPasswordHash)

	mockQueue := &queue.MockQueue{}
	mockQueue.On("UnsafePush", mock.AnythingOfType("[]uint8")).Return(nil)

	cst.encoder = mockEncoder
	cst.queue = mockQueue
}

func (cst *createUserSuite) TestCreateUserSuccess() {
	mockStore := &user.MockStore{}
	mockStore.On(
		"CreateUser",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("User"),
	).Return(test.UserID, nil)

	cst.encoder.(*password.MockEncoder).On("ValidatePassword", test.UserPassword).Return(nil)

	service := user.NewService(mockStore, cst.encoder, cst.queue)

	_, err := service.CreateUser(context.Background(), test.UserName, test.UserEmail, test.UserPassword)
	assert.Nil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenStoreCallFails() {
	mockStore := &user.MockStore{}
	mockStore.On(
		"CreateUser",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("User"),
	).Return("", liberr.WithArgs(errors.New("failed to save new user")))

	cst.encoder.(*password.MockEncoder).On("ValidatePassword", test.UserPassword).Return(nil)

	service := user.NewService(mockStore, cst.encoder, cst.queue)

	_, err := service.CreateUser(context.Background(), test.UserName, test.UserEmail, test.UserPassword)
	assert.NotNil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenInputIsInvalid() {
	testCases := map[string]struct {
		input func() (string, string, string)
		err   error
	}{
		"test failure when name is empty": {
			input: func() (string, string, string) {
				return test.EmptyString, test.UserEmail, test.UserPassword
			},
			err: errors.New("name cannot be empty"),
		},
		"test failure when email is empty": {
			input: func() (string, string, string) {
				return test.UserName, test.EmptyString, test.UserPassword
			},
			err: errors.New("email cannot be empty"),
		},
		"test failure when pass is empty": {
			input: func() (string, string, string) {
				return test.UserName, test.UserEmail, test.EmptyString
			},
			err: errors.New("password cannot be empty"),
		},
		"test failure when password is invalid": {
			input: func() (string, string, string) {
				cst.encoder.(*password.MockEncoder).On(
					"ValidatePassword",
					test.UserPasswordInvalid,
				).Return(liberr.WithArgs(errors.New("invalid password")))
				return test.UserName, test.UserEmail, test.UserPasswordInvalid
			},
			err: errors.New("invalid password"),
		},
	}

	for name, testCase := range testCases {
		cst.T().Run(name, func(t *testing.T) {
			service := user.NewService(&user.MockStore{}, cst.encoder, &queue.MockQueue{})

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
	mockStore := &user.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), test.UserEmail).Return(user.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"VerifyPassword",
		test.UserPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil)

	service := user.NewService(mockStore, mockEncoder, &queue.MockQueue{})

	_, err := service.GetUserID(context.Background(), test.UserEmail, test.UserPassword)
	require.NoError(t, err)
}

func TestGetUserIDFailureWhenStoreCallsFails(t *testing.T) {
	mockStore := &user.MockStore{}
	mockStore.On(
		"GetUser",
		mock.AnythingOfType("*context.timerCtx"),
		test.UserEmail,
	).Return(user.User{}, liberr.WithArgs(errors.New("failed to get user")))

	service := user.NewService(mockStore, &password.MockEncoder{}, &queue.MockQueue{})

	_, err := service.GetUserID(context.Background(), test.UserEmail, test.UserPassword)
	require.Error(t, err)
}

func TestGetUserIDFailureWhenPasswordVerificationFails(t *testing.T) {
	mockStore := &user.MockStore{}
	mockStore.On(
		"GetUser",
		mock.AnythingOfType("*context.timerCtx"),
		test.UserEmail,
	).Return(user.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"VerifyPassword",
		test.UserPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(liberr.WithArgs(errors.New("invalid credentials")))

	service := user.NewService(mockStore, mockEncoder, &queue.MockQueue{})

	_, err := service.GetUserID(context.Background(), test.UserEmail, test.UserPassword)
	require.Error(t, err)
}

func TestUpdatePasswordSuccess(t *testing.T) {
	mockStore := &user.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), test.UserEmail).Return(user.User{}, nil)
	mockStore.On(
		"UpdatePassword",
		mock.AnythingOfType("*context.timerCtx"),
		mock.AnythingOfType("string"), test.UserPasswordHash, test.UserPasswordSalt,
	).Return(int64(1), nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(test.UserPasswordSalt, nil)
	mockEncoder.On("GenerateKey", test.UserPasswordNew, test.UserPasswordSalt).Return(test.UserPasswordKey)
	mockEncoder.On("EncodeKey", test.UserPasswordKey).Return(test.UserPasswordHash)
	mockEncoder.On("VerifyPassword",
		test.UserPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil)
	mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)

	service := user.NewService(mockStore, mockEncoder, &queue.MockQueue{})

	err := service.UpdatePassword(context.Background(), test.UserEmail, test.UserPassword, test.UserPasswordNew)
	require.NoError(t, err)
}

func TestUpdatePasswordFailure(t *testing.T) {
	testCases := map[string]struct {
		store   func() user.Store
		encoder func() password.Encoder
	}{
		"test failure when new password does not match spec": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.timerCtx"),
					test.UserEmail,
				).Return(user.User{}, liberr.WithArgs(errors.New("failed to get user")))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On(
					"ValidatePassword",
					test.UserPasswordNew,
				).Return(liberr.WithArgs(errors.New("invalid password")))

				return mockEncoder
			},
		},
		"test failure when get user fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.timerCtx"),
					test.UserEmail,
				).Return(user.User{}, liberr.WithArgs(errors.New("failed to get user")))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)

				return mockEncoder
			},
		},
		"test failure when password verification fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.timerCtx"),
					test.UserEmail,
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
				).Return(liberr.WithArgs(errors.New("invalid credentials")))

				return mockEncoder
			},
		},
		"test failure when generate salt fails fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.timerCtx"),
					test.UserEmail,
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
					test.UserPasswordSalt,
					liberr.WithArgs(errors.New("failed to generate salt")),
				)

				return mockEncoder
			},
		},
		"test failure when generate store call fails": {
			store: func() user.Store {
				mockStore := &user.MockStore{}
				mockStore.On(
					"GetUser",
					mock.AnythingOfType("*context.timerCtx"),
					test.UserEmail,
				).Return(user.User{}, nil)
				mockStore.On(
					"UpdatePassword",
					mock.AnythingOfType("*context.timerCtx"),
					mock.AnythingOfType("string"),
					test.UserPasswordHash,
					test.UserPasswordSalt,
				).Return(int64(0), liberr.WithArgs(errors.New("failed to update password")))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", test.UserPasswordNew).Return(nil)
				mockEncoder.On("GenerateSalt").Return(test.UserPasswordSalt, nil)
				mockEncoder.On("GenerateKey", test.UserPasswordNew, test.UserPasswordSalt).Return(test.UserPasswordKey)
				mockEncoder.On("EncodeKey", test.UserPasswordKey).Return(test.UserPasswordHash)
				mockEncoder.On(
					"VerifyPassword",
					test.UserPassword,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				return mockEncoder
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := user.NewService(testCase.store(), testCase.encoder(), &queue.MockQueue{})

			err := service.UpdatePassword(context.Background(), test.UserEmail, test.UserPassword, test.UserPasswordNew)
			require.Error(t, err)
		})
	}
}
