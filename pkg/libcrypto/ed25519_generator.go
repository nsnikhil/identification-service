package libcrypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/ssh"
	"identification-service/pkg/liberr"
)

type Ed25519Generator interface {
	Generate() (ed25519.PublicKey, ed25519.PrivateKey, error)
	FromEncodedPem(encodedPem string) (ed25519.PublicKey, ed25519.PrivateKey, error)
}

//TODO: RENAME
type ed25519KeyGenerator struct {
}

func (eg *ed25519KeyGenerator) Generate() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pubKey, priKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, liberr.WithOp("Ed25519Generator.Generate", err)
	}

	return pubKey, priKey, err
}

func (eg *ed25519KeyGenerator) FromEncodedPem(encodedPem string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pem, err := base64.RawStdEncoding.DecodeString(encodedPem)
	if err != nil {
		return nil, nil, liberr.WithOp("Ed25519Generator.FromEncodedPem", err)
	}

	privateKey, err := ssh.ParseRawPrivateKey(pem)
	if err != nil {
		return nil, nil, liberr.WithOp("Ed25519Generator.FromEncodedPem", err)
	}

	priKey, ok := privateKey.(*ed25519.PrivateKey)
	if !ok {
		return nil, nil, liberr.WithOp(
			"Ed25519Generator.FromEncodedPem",
			fmt.Errorf("invalid ed25519 key %v", pem),
		)
	}

	return priKey.Public().(ed25519.PublicKey), *priKey, nil
}

func NewKeyGenerator() Ed25519Generator {
	return &ed25519KeyGenerator{}
}
