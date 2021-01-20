package config

import "github.com/stretchr/testify/mock"

type ClientConfig interface {
	Strategies() map[string]bool
}

type appClientConfig struct {
	strategies map[string]bool
}

func newClientConfig() ClientConfig {
	return appClientConfig{
		strategies: toSet(getStringSlice("STRATEGIES")),
	}
}

func toSet(values []string) map[string]bool {
	res := make(map[string]bool)
	for _, value := range values {
		res[value] = true
	}
	return res
}

func (sc appClientConfig) Strategies() map[string]bool {
	return sc.strategies
}

type MockClientConfig struct {
	mock.Mock
}

func (mock *MockClientConfig) Strategies() map[string]bool {
	args := mock.Called()
	return args.Get(0).(map[string]bool)
}
