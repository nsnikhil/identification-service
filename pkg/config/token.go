package config

type TokenConfig struct {
	audience          string
	issuer            string
	encodedSigningKey string
}

func newTokenConfig() TokenConfig {
	return TokenConfig{
		audience:          getString("TOKEN_AUDIENCE"),
		issuer:            getString("TOKEN_ISSUER"),
		encodedSigningKey: getString("ENCODED_TOKEN_SIGNING_KEY"),
	}
}

func (tc TokenConfig) Audience() string {
	return tc.audience
}

func (tc TokenConfig) Issuer() string {
	return tc.issuer
}

func (tc TokenConfig) EncodedSigningKey() string {
	return tc.encodedSigningKey
}
