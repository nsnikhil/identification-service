package internal

import (
	"errors"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"time"
)

type User struct {
	id string

	name  string
	email string

	passwordHash string
	passwordSalt []byte

	createdAt time.Time
	updatedAt time.Time
}

func (u User) ID() string {
	return u.id
}

// TODO: REMOVE THIS GETTER
func (u User) PasswordHash() string {
	return u.passwordHash
}

// TODO: REMOVE THIS GETTER
func (u User) PasswordSalt() []byte {
	return u.passwordSalt
}

//TODO: CHANGE TO BUILDER PATTERN
func NewUser(encoder password.Encoder, name, email, password string) (User, error) {
	err := validateArgs(encoder, name, email, password)
	if err != nil {
		return User{}, liberr.WithArgs(liberr.Operation("User.NewUser"), liberr.ValidationError, err)
	}

	salt, err := encoder.GenerateSalt()
	if err != nil {
		return User{}, liberr.WithOp("User.NewUser", err)
	}

	key := encoder.GenerateKey(password, salt)
	hash := encoder.EncodeKey(key)

	return User{
		name:         name,
		email:        email,
		passwordSalt: salt,
		passwordHash: hash,
	}, nil
}

func validateArgs(encoder password.Encoder, name, email, password string) error {
	if len(name) == 0 {
		return errors.New("name cannot be empty")
	}

	if len(email) == 0 {
		return errors.New("email cannot be empty")
	}

	if len(password) == 0 {
		return errors.New("password cannot be empty")
	}

	return encoder.ValidatePassword(password)
}
