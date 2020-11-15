package libcrypto

import (
	"crypto/ed25519"
	"github.com/stretchr/testify/mock"
)

type MockEd25519Generator struct {
	mock.Mock
}

func (mock *MockEd25519Generator) Generate() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	args := mock.Called()
	return args.Get(0).(ed25519.PublicKey), args.Get(1).(ed25519.PrivateKey), args.Error(2)
}

func (mock *MockEd25519Generator) FromEncodedPem(encodedPem string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	args := mock.Called(encodedPem)
	return args.Get(0).(ed25519.PublicKey), args.Get(1).(ed25519.PrivateKey), args.Error(2)
}
