package config

type PasswordConfig struct {
	saltLength, iterations, keyLength int
}

func newPasswordConfig() PasswordConfig {
	return PasswordConfig{
		saltLength: getInt("PASSWORD_HASH_SALT_LENGTH"),
		iterations: getInt("PASSWORD_HASH_ITERATIONS"),
		keyLength:  getInt("PASSWORD_HASH_KEY_LENGTH"),
	}
}

func (pc PasswordConfig) SaltLength() int {
	return pc.saltLength
}

func (pc PasswordConfig) Iterations() int {
	return pc.iterations
}

func (pc PasswordConfig) KeyLength() int {
	return pc.keyLength
}
