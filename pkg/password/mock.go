package password

import "github.com/stretchr/testify/mock"

type MockEncoder struct {
	mock.Mock
}

func (mock *MockEncoder) GenerateSalt() ([]byte, error) {
	args := mock.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (mock *MockEncoder) GenerateKey(password string, salt []byte) []byte {
	args := mock.Called(password, salt)
	return args.Get(0).([]byte)
}

func (mock *MockEncoder) EncodeKey(key []byte) string {
	args := mock.Called(key)
	return args.String(0)
}

func (mock *MockEncoder) VerifyPassword(password, userPasswordHash string, userPasswordSalt []byte) error {
	args := mock.Called(password, userPasswordHash, userPasswordSalt)
	return args.Error(0)
}

func (mock *MockEncoder) ValidatePassword(password string) error {
	args := mock.Called(password)
	return args.Error(0)
}
