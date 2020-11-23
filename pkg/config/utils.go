package config

import (
	"github.com/spf13/viper"
	"strings"
)

func getString(config string, defaultVal ...string) string {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	return viper.GetString(config)
}

func getStringSlice(config string, defaultVal ...string) []string {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	return strings.Split(viper.GetString(config), ",")
}

func getStringMap(config string, defaultVal ...string) map[string]string {
	if len(defaultVal) > 0 {
		viper.SetDefault(config, defaultVal[0])
	}

	pairs := strings.Split(viper.GetString(config), ",")

	res := make(map[string]string)

	for _, pair := range pairs {
		data := strings.Split(pair, ":")
		res[data[0]] = data[1]
	}

	return res
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
