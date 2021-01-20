package config

import "github.com/stretchr/testify/mock"

type PasswordConfig interface {
	SaltLength() int
	Iterations() int
	KeyLength() int
}

type appPasswordConfig struct {
	saltLength, iterations, keyLength int
}

func newPasswordConfig() PasswordConfig {
	return appPasswordConfig{
		saltLength: getInt("PASSWORD_HASH_SALT_LENGTH"),
		iterations: getInt("PASSWORD_HASH_ITERATIONS"),
		keyLength:  getInt("PASSWORD_HASH_KEY_LENGTH"),
	}
}

func (pc appPasswordConfig) SaltLength() int {
	return pc.saltLength
}

func (pc appPasswordConfig) Iterations() int {
	return pc.iterations
}

func (pc appPasswordConfig) KeyLength() int {
	return pc.keyLength
}

type MockPasswordConfig struct {
	mock.Mock
}

func (mock *MockPasswordConfig) SaltLength() int {
	args := mock.Called()
	return args.Int(0)
}

func (mock *MockPasswordConfig) Iterations() int {
	args := mock.Called()
	return args.Int(0)
}

func (mock *MockPasswordConfig) KeyLength() int {
	args := mock.Called()
	return args.Int(0)
}
