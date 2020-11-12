package config

import (
	"fmt"
)

type HTTPServerConfig struct {
	host             string
	port             int
	readTimoutInSec  int
	writeTimoutInSec int
}

func newHTTPServerConfig() HTTPServerConfig {
	return HTTPServerConfig{
		host:             getString("HTTP_SERVER_HOST"),
		port:             getInt("HTTP_SERVER_PORT"),
		readTimoutInSec:  getInt("HTTP_SERVER_READ_TIMEOUT_IN_SEC"),
		writeTimoutInSec: getInt("HTTP_SERVER_WRITE_TIMEOUT_IN_SEC"),
	}
}

func (sc HTTPServerConfig) Address() string {
	return fmt.Sprintf(":%d", sc.port)
}

func (sc HTTPServerConfig) ReadTimeout() int {
	return sc.readTimoutInSec
}

func (sc HTTPServerConfig) WriteTimeout() int {
	return sc.readTimoutInSec
}
