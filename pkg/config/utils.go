package config

import "github.com/spf13/viper"

func getString(config string, defaultVal ...string) string {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	return viper.GetString(config)
}

func getInt(config string, defaultVal ...int) int {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	return viper.GetInt(config)
}

func getFloat(config string, defaultVal ...int) float64 {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	return viper.GetFloat64(config)
}

func getBool(config string, defaultVal ...bool) bool {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	return viper.GetBool(config)
}
