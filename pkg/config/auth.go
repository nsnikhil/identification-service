package config

type AuthConfig struct {
	userName string
	password string
}

func (ac AuthConfig) UserName() string {
	return ac.userName
}

func (ac AuthConfig) Password() string {
	return ac.password
}

func newAuthConfig() AuthConfig {
	return AuthConfig{
		userName: getString("BASIC_AUTH_USER_NAME"),
		password: getString("BASIC_AUTH_PASSWORD"),
	}
}
