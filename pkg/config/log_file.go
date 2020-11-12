package config

import "path"

type LogFileConfig struct {
	name               string
	dir                string
	maxSizeInMb        int
	maxBackups         int
	maxAge             int
	withLocalTimeStamp bool
}

func (lfc LogFileConfig) GetFileName() string {
	return lfc.name
}

func (lfc LogFileConfig) GetFileDir() string {
	return lfc.dir
}

func (lfc LogFileConfig) GetFilePath() string {
	return path.Join(lfc.dir, lfc.name)
}

func (lfc LogFileConfig) GetFileMaxSizeInMb() int {
	return lfc.maxSizeInMb
}

func (lfc LogFileConfig) GetFileMaxBackups() int {
	return lfc.maxBackups
}

func (lfc LogFileConfig) GetFileMaxAge() int {
	return lfc.maxAge
}

func (lfc LogFileConfig) GetFileWithLocalTimeStamp() bool {
	return lfc.withLocalTimeStamp
}

func newLogFileConfig() LogFileConfig {
	return LogFileConfig{
		name:               getString("LOG_FILE_NAME"),
		dir:                getString("LOG_FILE_DIR"),
		maxSizeInMb:        getInt("LOG_FILE_MAX_SIZE_IN_MB"),
		maxBackups:         getInt("LOG_FILE_MAX_BACKUPS"),
		maxAge:             getInt("LOG_FILE_MAX_AGE"),
		withLocalTimeStamp: getBool("LOG_FILE_WITH_LOCAL_TIME_STAMP"),
	}
}
