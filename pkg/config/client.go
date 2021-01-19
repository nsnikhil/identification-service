package config

type ClientConfig struct {
	strategies map[string]bool
}

func newClientConfig() ClientConfig {
	return ClientConfig{
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

func (sc ClientConfig) Strategies() map[string]bool {
	return sc.strategies
}
