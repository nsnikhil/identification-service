package user_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/password"
	"identification-service/pkg/producer"
	"identification-service/pkg/test"
	"identification-service/pkg/user"
	"testing"
)

type createUserSuite struct {
	cfg      config.KafkaConfig
	encoder  password.Encoder
	producer producer.Producer
	suite.Suite
}

func (cst *createUserSuite) SetupSuite() {
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)

	mockProducer := &producer.MockProducer{}
	mockProducer.On("Produce", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
		Return(int32(0), int64(0), nil)

	mockKafkaConfig := &config.MockKafkaConfig{}
	mockKafkaConfig.On("SignUpTopicName").Return("sign-up")
	mockKafkaConfig.On("UpdatePasswordTopicName").Return("update-password")

	cst.encoder = mockEncoder
	cst.producer = mockProducer
	cst.cfg = mockKafkaConfig
}

func (cst *createUserSuite) TestCreateUserSuccess() {
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	mockStore := &user.MockStore{}
	mockStore.On(
		"CreateUser",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("User"),
	).Return(test.NewUUID(), nil)

	//TODO: OVERRIDING GLOBAL ENCODER (REFACTOR)
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey, nil)
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	service := user.NewService(cst.cfg, mockStore, mockEncoder, cst.producer)

	_, err := service.CreateUser(context.Background(), test.RandString(8), test.NewEmail(), userPassword)
	assert.Nil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenStoreCallFails() {
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPassword := test.NewPassword()

	mockStore := &user.MockStore{}
	mockStore.On(
		"CreateUser",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("User"),
	).Return("", errors.New("failed to save new user"))

	//TODO: OVERRIDING GLOBAL ENCODER (REFACTOR)
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateKey", userPassword, passwordSalt).Return(passwordKey, nil)
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("ValidatePassword", userPassword).Return(nil)

	service := user.NewService(cst.cfg, mockStore, mockEncoder, cst.producer)

	_, err := service.CreateUser(context.Background(), test.RandString(8), test.NewEmail(), userPassword)
	assert.NotNil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenInputIsInvalid() {
	invalidPassword := test.RandString(12)

	testCases := map[string]struct {
		input func() (string, string, string)
		err   error
	}{
		"test failure when name is empty": {
			input: func() (string, string, string) {
				return test.EmptyString, test.NewEmail(), test.NewPassword()
			},
			err: errors.New("name cannot be empty"),
		},
		"test failure when email is empty": {
			input: func() (string, string, string) {
				return test.RandString(8), test.EmptyString, test.NewPassword()
			},
			err: errors.New("email cannot be empty"),
		},
		"test failure when pass is empty": {
			input: func() (string, string, string) {
				return test.RandString(8), test.NewEmail(), test.EmptyString
			},
			err: errors.New("password cannot be empty"),
		},
		"test failure when password is invalid": {
			input: func() (string, string, string) {
				cst.encoder.(*password.MockEncoder).On(
					"ValidatePassword",
					invalidPassword,
				).Return(errors.New("invalid password"))
				return test.RandString(8), test.NewEmail(), invalidPassword
			},
			err: errors.New("invalid password"),
		},
	}

	for name, testCase := range testCases {
		cst.T().Run(name, func(t *testing.T) {

			service := user.NewService(cst.cfg, &user.MockStore{}, cst.encoder, &producer.MockProducer{})

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
	userEmail := test.NewEmail()
	userPassword := test.NewPassword()

	mockStore := &user.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.emptyCtx"), userEmail).Return(user.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"VerifyPassword",
		userPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil)

	service := user.NewService(&config.MockKafkaConfig{}, mockStore, mockEncoder, &producer.MockProducer{})

	_, err := service.GetUserID(context.Background(), userEmail, userPassword)
	require.NoError(t, err)
}

func TestGetUserIDFailureWhenStoreCallsFails(t *testing.T) {
	userEmail := test.NewEmail()

	mockStore := &user.MockStore{}
	mockStore.On(
		"GetUser",
		mock.AnythingOfType("*context.emptyCtx"),
		userEmail,
	).Return(user.User{}, errors.New("failed to get user"))

	service := user.NewService(&config.MockKafkaConfig{}, mockStore, &password.MockEncoder{}, &producer.MockProducer{})

	_, err := service.GetUserID(context.Background(), userEmail, test.NewPassword())
	require.Error(t, err)
}

func TestGetUserIDFailureWhenPasswordVerificationFails(t *testing.T) {
	userEmail := test.NewEmail()
	userPassword := test.NewPassword()

	mockStore := &user.MockStore{}
	mockStore.On(
		"GetUser",
		mock.AnythingOfType("*context.emptyCtx"),
		userEmail,
	).Return(user.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On(
		"VerifyPassword",
		userPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(errors.New("invalid credentials"))

	service := user.NewService(&config.MockKafkaConfig{}, mockStore, mockEncoder, &producer.MockProducer{})

	_, err := service.GetUserID(context.Background(), userEmail, userPassword)
	require.Error(t, err)
}

func TestUpdatePasswordSuccess(t *testing.T) {
	userEmail := test.NewEmail()
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	userPasswordNew := test.NewPassword()
	userPassword := test.NewPassword()

	mockStore := &user.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.emptyCtx"), userEmail).Return(user.User{}, nil)
	mockStore.On(
		"UpdatePassword",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("string"), passwordHash, passwordSalt,
	).Return(int64(1), nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
	mockEncoder.On("GenerateKey", userPasswordNew, passwordSalt).Return(passwordKey)
	mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
	mockEncoder.On("VerifyPassword",
		userPassword,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("[]uint8"),
	).Return(nil)
	mockEncoder.On("ValidatePassword", userPasswordNew).Return(nil)

	mockProducer := &producer.MockProducer{}
	mockProducer.On("Produce", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
		Return(int32(0), int64(0), nil)

	mockKafkaConfig := &config.MockKafkaConfig{}
	mockKafkaConfig.On("UpdatePasswordTopicName").Return("update-password")

	service := user.NewService(mockKafkaConfig, mockStore, mockEncoder, mockProducer)

	err := service.UpdatePassword(context.Background(), userEmail, userPassword, userPasswordNew)
	require.NoError(t, err)
}

func TestUpdatePasswordFailure(t *testing.T) {
	userEmail := test.NewEmail()
	passwordSalt := test.RandBytes(86)
	passwordKey := test.RandBytes(32)
	passwordHash := test.RandString(44)
	invalidPassword := test.RandString(12)
	userPassword := test.NewPassword()
	userPasswordNew := test.NewPassword()

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
					invalidPassword,
				).Return(errors.New("invalid password"))

				return mockEncoder
			},
			newPassword: invalidPassword,
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
				mockEncoder.On("ValidatePassword", userPasswordNew).Return(nil)

				return mockEncoder
			},
			newPassword: userPasswordNew,
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
				mockEncoder.On("ValidatePassword", userPasswordNew).Return(nil)
				mockEncoder.On(
					"VerifyPassword",
					userPassword,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				mockEncoder.On("GenerateSalt").Return(
					passwordSalt,
					errors.New("failed to generate salt"),
				)

				return mockEncoder
			},
			newPassword: userPasswordNew,
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
				mockEncoder.On("ValidatePassword", userPasswordNew).Return(nil)
				mockEncoder.On("GenerateSalt").Return(passwordSalt, nil)
				mockEncoder.On("GenerateKey", userPasswordNew, passwordSalt).Return(passwordKey)
				mockEncoder.On("EncodeKey", passwordKey).Return(passwordHash)
				mockEncoder.On(
					"VerifyPassword",
					userPassword,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("[]uint8"),
				).Return(nil)

				return mockEncoder
			},
			newPassword: userPasswordNew,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			mockKafkaConfig := &config.MockKafkaConfig{}
			mockKafkaConfig.On("UpdatePasswordTopicName").Return("update-password")

			service := user.NewService(mockKafkaConfig, testCase.store(), testCase.encoder(), &producer.MockProducer{})

			err := service.UpdatePassword(context.Background(), userEmail, userPassword, testCase.newPassword)
			require.Error(t, err)
		})
	}
}
