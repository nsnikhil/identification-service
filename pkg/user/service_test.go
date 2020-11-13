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
	"identification-service/pkg/user"
	"identification-service/pkg/user/internal"
	"testing"
)

const (
	name         = "Test Name"
	email        = "test@test.com"
	userPassword = "Password@1234"
	newPassword  = "NewPassword@1234"

	invalidPasswordOne = "password@1234"

	userID = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
)

var salt = []byte{90, 20, 247, 194, 220, 48, 153, 58, 158, 103, 9, 17, 243, 24, 179, 254, 88, 59, 161, 81, 216, 8, 126, 122, 102, 151, 200, 12, 134, 118, 146, 197, 193, 248, 117, 57, 127, 137, 112, 233, 116, 50, 128, 84, 127, 93, 180, 23, 81, 69, 245, 183, 45, 57, 51, 125, 9, 46, 200, 175, 97, 49, 11, 0, 40, 228, 186, 60, 177, 43, 69, 52, 168, 195, 69, 101, 21, 245, 62, 131, 252, 96, 240, 154, 251, 2}
var key = []byte{34, 179, 107, 154, 0, 94, 48, 1, 134, 44, 128, 127, 254, 17, 124, 248, 69, 96, 196, 174, 146, 255, 131, 91, 94, 143, 105, 33, 230, 157, 77, 243}
var hash = "IrNrmgBeMAGGLIB//hF8+EVgxK6S/4NbXo9pIeadTfM="

type createUserSuite struct {
	encoder password.Encoder
	queue   queue.Queue
	suite.Suite
}

func (cst *createUserSuite) SetupSuite() {
	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(salt, nil)
	mockEncoder.On("GenerateKey", userPassword, salt).Return(key)
	mockEncoder.On("EncodeKey", key).Return(hash)

	mockQueue := &queue.MockQueue{}
	mockQueue.On("UnsafePush", mock.AnythingOfType("[]uint8")).Return(nil)

	cst.encoder = mockEncoder
	cst.queue = mockQueue
}

func (cst *createUserSuite) TestCreateUserSuccess() {
	mockStore := &internal.MockStore{}
	mockStore.On("CreateUser", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.User")).Return(userID, nil)

	cst.encoder.(*password.MockEncoder).On("ValidatePassword", userPassword).Return(nil)

	service := user.NewInternalService(mockStore, cst.encoder, cst.queue)

	_, err := service.CreateUser(context.Background(), name, email, userPassword)
	assert.Nil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenStoreCallFails() {
	mockStore := &internal.MockStore{}
	mockStore.On("CreateUser", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.User")).Return("", liberr.WithArgs(errors.New("failed to save new user")))

	cst.encoder.(*password.MockEncoder).On("ValidatePassword", userPassword).Return(nil)

	service := user.NewInternalService(mockStore, cst.encoder, cst.queue)

	_, err := service.CreateUser(context.Background(), name, email, userPassword)
	assert.NotNil(cst.T(), err)
}

func (cst *createUserSuite) TestCreateFailureWhenInputIsInvalid() {
	testCases := map[string]struct {
		input func() (string, string, string)
		err   error
	}{
		"test failure when name is empty": {
			input: func() (string, string, string) {
				return "", email, userPassword
			},
			err: errors.New("name cannot be empty"),
		},
		"test failure when email is empty": {
			input: func() (string, string, string) {
				return name, "", userPassword
			},
			err: errors.New("email cannot be empty"),
		},
		"test failure when pass is empty": {
			input: func() (string, string, string) {
				return name, email, ""
			},
			err: errors.New("password cannot be empty"),
		},
		"test failure when password is invalid": {
			input: func() (string, string, string) {
				cst.encoder.(*password.MockEncoder).On("ValidatePassword", invalidPasswordOne).Return(liberr.WithArgs(errors.New("invalid password")))
				return name, email, invalidPasswordOne
			},
			err: errors.New("invalid password"),
		},
	}

	for name, testCase := range testCases {
		cst.T().Run(name, func(t *testing.T) {
			service := user.NewInternalService(&internal.MockStore{}, cst.encoder, &queue.MockQueue{})

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
	mockStore := &internal.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("VerifyPassword", userPassword, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	service := user.NewInternalService(mockStore, mockEncoder, &queue.MockQueue{})

	_, err := service.GetUserID(context.Background(), email, userPassword)
	require.NoError(t, err)
}

func TestGetUserIDFailureWhenStoreCallsFails(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, liberr.WithArgs(errors.New("failed to get user")))

	service := user.NewInternalService(mockStore, &password.MockEncoder{}, &queue.MockQueue{})

	_, err := service.GetUserID(context.Background(), email, userPassword)
	require.Error(t, err)
}

func TestGetUserIDFailureWhenPasswordVerificationFails(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("VerifyPassword", userPassword, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(liberr.WithArgs(errors.New("invalid credentials")))

	service := user.NewInternalService(mockStore, mockEncoder, &queue.MockQueue{})

	_, err := service.GetUserID(context.Background(), email, userPassword)
	require.Error(t, err)
}

func TestUpdatePasswordSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, nil)
	mockStore.On("UpdatePassword", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), hash, salt).Return(int64(1), nil)

	mockEncoder := &password.MockEncoder{}
	mockEncoder.On("GenerateSalt").Return(salt, nil)
	mockEncoder.On("GenerateKey", newPassword, salt).Return(key)
	mockEncoder.On("EncodeKey", key).Return(hash)
	mockEncoder.On("VerifyPassword", userPassword, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	mockEncoder.On("ValidatePassword", newPassword).Return(nil)

	service := user.NewInternalService(mockStore, mockEncoder, &queue.MockQueue{})

	err := service.UpdatePassword(context.Background(), email, userPassword, newPassword)
	require.NoError(t, err)
}

func TestUpdatePasswordFailure(t *testing.T) {
	testCases := map[string]struct {
		store   func() internal.Store
		encoder func() password.Encoder
	}{
		"test failure when new password does not match spec": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, liberr.WithArgs(errors.New("failed to get user")))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", newPassword).Return(liberr.WithArgs(errors.New("invalid password")))

				return mockEncoder
			},
		},
		"test failure when get user fails": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, liberr.WithArgs(errors.New("failed to get user")))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", newPassword).Return(nil)

				return mockEncoder
			},
		},
		"test failure when password verification fails": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, nil)

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", newPassword).Return(nil)
				mockEncoder.On("VerifyPassword", userPassword, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(liberr.WithArgs(errors.New("invalid credentials")))

				return mockEncoder
			},
		},
		"test failure when generate salt fails fails": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, nil)

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", newPassword).Return(nil)
				mockEncoder.On("VerifyPassword", userPassword, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
				mockEncoder.On("GenerateSalt").Return(salt, liberr.WithArgs(errors.New("failed to generate salt")))

				return mockEncoder
			},
		},
		"test failure when generate store call fails": {
			store: func() internal.Store {
				mockStore := &internal.MockStore{}
				mockStore.On("GetUser", mock.AnythingOfType("*context.timerCtx"), email).Return(internal.User{}, nil)
				mockStore.On("UpdatePassword", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("string"), hash, salt).Return(int64(0), liberr.WithArgs(errors.New("failed to update password")))

				return mockStore
			},
			encoder: func() password.Encoder {
				mockEncoder := &password.MockEncoder{}
				mockEncoder.On("ValidatePassword", newPassword).Return(nil)
				mockEncoder.On("GenerateSalt").Return(salt, nil)
				mockEncoder.On("GenerateKey", newPassword, salt).Return(key)
				mockEncoder.On("EncodeKey", key).Return(hash)
				mockEncoder.On("VerifyPassword", userPassword, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

				return mockEncoder
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			service := user.NewInternalService(testCase.store(), testCase.encoder(), &queue.MockQueue{})

			err := service.UpdatePassword(context.Background(), email, userPassword, newPassword)
			require.Error(t, err)
		})
	}
}
