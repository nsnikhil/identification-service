package config

import "strings"

type LogConfig struct {
	level string
	sinks []string
}

func (lc LogConfig) Level() string {
	return lc.level
}

func (lc LogConfig) Sinks() []string {
	return lc.sinks
}

func newLogConfig() LogConfig {
	return LogConfig{
		level: getString("LOG_LEVEL"),
		sinks: strings.Split(getString("LOG_SINK"), ","),
	}
}
