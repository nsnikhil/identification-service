package user

import (
	"errors"
	"fmt"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/util"
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

type Builder struct {
	id string

	name  string
	email string

	passwordHash string
	passwordSalt []byte

	createdAt time.Time
	updatedAt time.Time

	encoder password.Encoder

	err error
}

func (b *Builder) ID(id string) *Builder {
	if b.err != nil {
		return b
	}

	if !util.IsValidUUID(id) {
		b.err = fmt.Errorf("invalid user id %s", id)
		return b
	}

	b.id = id
	return b
}

func (b *Builder) Name(name string) *Builder {
	if b.err != nil {
		return b
	}

	if len(name) == 0 {
		b.err = errors.New("name cannot be empty")
		return b
	}

	b.name = name
	return b
}

func (b *Builder) Email(email string) *Builder {
	if b.err != nil {
		return b
	}

	if len(email) == 0 {
		b.err = errors.New("email cannot be empty")
		return b
	}

	b.email = email
	return b
}

func (b *Builder) Password(password string) *Builder {
	if b.err != nil {
		return b
	}

	if len(password) == 0 {
		b.err = errors.New("password cannot be empty")
		return b
	}

	err := b.encoder.ValidatePassword(password)
	if err != nil {
		b.err = err
		return b
	}

	salt, err := b.encoder.GenerateSalt()
	if err != nil {
		b.err = err
		return b
	}

	key := b.encoder.GenerateKey(password, salt)
	hash := b.encoder.EncodeKey(key)

	b.passwordSalt = salt
	b.passwordHash = hash
	return b
}

func (b *Builder) PasswordHash(passwordHash string) *Builder {
	if b.err != nil {
		return b
	}

	if len(passwordHash) == 0 {
		b.err = errors.New("password hash cannot be empty")
		return b
	}

	b.passwordHash = passwordHash
	return b
}

func (b *Builder) PasswordSalt(passwordSalt []byte) *Builder {
	if b.err != nil {
		return b
	}

	if passwordSalt == nil || len(passwordSalt) == 0 {
		b.err = errors.New("password salt cannot be empty")
		return b
	}

	b.passwordSalt = passwordSalt
	return b
}

func (b *Builder) CreatedAt(createdAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	b.createdAt = createdAt
	return b
}

func (b *Builder) UpdatedAt(updatedAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	b.updatedAt = updatedAt
	return b
}

func (b *Builder) Build() (User, error) {
	if b.err != nil {
		return User{}, liberr.WithArgs(liberr.Operation("Builder.Build"), liberr.ValidationError, b.err)
	}

	//TODO: ADD VALIDATION AGAIN SINCE USER MIGHT NOT HAVE SET ANY REQUIRED FIELDS USING BUILDER PATTERN

	return User{
		id:           b.id,
		name:         b.name,
		email:        b.email,
		passwordSalt: b.passwordSalt,
		passwordHash: b.passwordHash,
		createdAt:    b.createdAt,
		updatedAt:    b.updatedAt,
	}, nil
}

func NewUserBuilder(encoder password.Encoder) *Builder {
	return &Builder{encoder: encoder}
}

//func NewUser(encoder password.Encoder, name, email, password string) (User, error) {
//	err := validateArgs(encoder, name, email, password)
//	if err != nil {
//		return User{}, liberr.WithArgs(liberr.Operation("User.NewUser"), liberr.ValidationError, err)
//	}
//
//	salt, err := encoder.GenerateSalt()
//	if err != nil {
//		return User{}, liberr.WithOp("User.NewUser", err)
//	}
//
//	key := encoder.GenerateKey(password, salt)
//	hash := encoder.EncodeKey(key)
//
//	return User{
//		name:         name,
//		email:        email,
//		passwordSalt: salt,
//		passwordHash: hash,
//	}, nil
//}

//func validateArgs(encoder password.Encoder, name, email, password string) error {
//	if len(name) == 0 {
//		return errors.New("name cannot be empty")
//	}
//
//	if len(email) == 0 {
//		return errors.New("email cannot be empty")
//	}
//
//	if len(password) == 0 {
//		return errors.New("password cannot be empty")
//	}
//
//	return encoder.ValidatePassword(password)
//}
