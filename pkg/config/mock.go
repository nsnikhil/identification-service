package config

import "github.com/stretchr/testify/mock"

type MockConfig struct {
	mock.Mock
}

func (mock *MockConfig) HTTPServerConfig() HTTPServerConfig {
	args := mock.Called()
	return args.Get(0).(HTTPServerConfig)
}

func (mock *MockConfig) DatabaseConfig() DatabaseConfig {
	args := mock.Called()
	return args.Get(0).(DatabaseConfig)
}

func (mock *MockConfig) LogConfig() LogConfig {
	args := mock.Called()
	return args.Get(0).(LogConfig)
}

func (mock *MockConfig) LogFileConfig() LogFileConfig {
	args := mock.Called()
	return args.Get(0).(LogFileConfig)
}

func (mock *MockConfig) Env() string {
	args := mock.Called()
	return args.Get(0).(string)
}

func (mock *MockConfig) MigrationConfig() MigrationConfig {
	args := mock.Called()
	return args.Get(0).(MigrationConfig)
}

func (mock *MockConfig) PasswordConfig() PasswordConfig {
	args := mock.Called()
	return args.Get(0).(PasswordConfig)
}

func (mock *MockConfig) TokenConfig() TokenConfig {
	args := mock.Called()
	return args.Get(0).(TokenConfig)
}

func (mock *MockConfig) AuthConfig() AuthConfig {
	args := mock.Called()
	return args.Get(0).(AuthConfig)
}

func (mock *MockConfig) CacheConfig() CacheConfig {
	args := mock.Called()
	return args.Get(0).(CacheConfig)
}

func (mock *MockConfig) ClientConfig() ClientConfig {
	args := mock.Called()
	return args.Get(0).(ClientConfig)
}

func (mock *MockConfig) QueueConfig() QueueConfig {
	args := mock.Called()
	return args.Get(0).(QueueConfig)
}
