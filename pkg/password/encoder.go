package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/nsnikhil/erx"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
	"identification-service/pkg/config"
	"io"
	"unicode"
)

type Encoder interface {
	GenerateSalt() ([]byte, error)
	GenerateKey(password string, salt []byte) []byte
	EncodeKey(key []byte) string

	ValidatePassword(password string) error
	VerifyPassword(password, userPasswordHash string, userPasswordSalt []byte) error
}

//TODO: RENAME
type pbkdfPasswordEncoder struct {
	saltLength, iterations, keyLength int
}

func (pe *pbkdfPasswordEncoder) GenerateSalt() ([]byte, error) {
	salt := make([]byte, pe.saltLength)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, erx.WithArgs(erx.Operation("Encoder.GenerateSalt"), err)
	}

	return salt, nil
}

func (pe *pbkdfPasswordEncoder) GenerateKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, pe.iterations, pe.keyLength, sha3.New512)
}

func (pe *pbkdfPasswordEncoder) EncodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func (pe *pbkdfPasswordEncoder) VerifyPassword(password, userPasswordHash string, userPasswordSalt []byte) error {
	passwordHash := pe.EncodeKey(pe.GenerateKey(password, userPasswordSalt))

	if userPasswordHash != passwordHash {
		return erx.WithArgs(erx.Operation("Encoder.VerifyPassword"), errors.New("invalid credentials"))
	}

	return nil
}

//TODO: PASSWORD SPEC SHOULD BE CONFIGURABLE
func (pe *pbkdfPasswordEncoder) ValidatePassword(password string) error {
	wrap := func(err error) error {
		return erx.WithArgs(erx.Operation("Encoder.ValidatePassword"), erx.ValidationError, err)
	}

	if len(password) < 8 {
		return wrap(errors.New("password must be at least 8 characters long"))
	}

	u, l, n, s := 0, 0, 0, 0

	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			n++
		case unicode.IsLower(c):
			l++
		case unicode.IsUpper(c):
			u++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			s++
		default:
			return wrap(fmt.Errorf("invalid character %c", c))
		}
	}

	isValid := u > 0 && l > 0 && n > 0 && s > 0

	if !isValid {
		return wrap(errors.New("password must have at least 1 number, 1 lower character, 1 upper character and 1 symbol"))
	}

	return nil
}

func NewEncoder(cfg config.PasswordConfig) Encoder {
	return &pbkdfPasswordEncoder{
		saltLength: cfg.SaltLength(),
		iterations: cfg.Iterations(),
		keyLength:  cfg.KeyLength(),
	}
}
