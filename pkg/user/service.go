package user

import (
	"context"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/queue"
)

//TODO: RENAME (APPEND USER IN THE NAME)
type Service interface {
	CreateUser(ctx context.Context, name, email, password string) (string, error)
	UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error
	GetUserID(ctx context.Context, email, password string) (string, error)
}

// TODO: RENAME
type userService struct {
	store   Store
	encoder password.Encoder
	queue   queue.Queue
}

func (us *userService) CreateUser(ctx context.Context, name, email, password string) (string, error) {
	wrap := func(err error) error { return liberr.WithOp("Service.SignUp", err) }

	user, err := NewUserBuilder(us.encoder).Name(name).Email(email).Password(password).Build()
	if err != nil {
		return "", wrap(err)
	}

	userID, err := us.store.CreateUser(ctx, user)
	if err != nil {
		return "", wrap(err)
	}

	go us.queue.UnsafePush([]byte(userID))

	return userID, nil
}

func (us *userService) GetUserID(ctx context.Context, email, password string) (string, error) {
	user, err := us.store.GetUser(ctx, email)
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("Service.GetUserID"), liberr.InvalidCredentialsError, err)
	}

	err = us.encoder.VerifyPassword(password, user.passwordHash, user.passwordSalt)
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("Service.GetUserID"), liberr.InvalidCredentialsError, err)
	}

	return user.id, nil
}

func (us *userService) UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error {
	wrap := func(err error) error { return liberr.WithOp("Service.UpdatePassword", err) }

	err := us.encoder.ValidatePassword(newPassword)
	if err != nil {
		return wrap(err)
	}

	id, err := us.GetUserID(ctx, email, oldPassword)
	if err != nil {
		return wrap(err)
	}

	salt, err := us.encoder.GenerateSalt()
	if err != nil {
		return wrap(err)
	}

	key := us.encoder.GenerateKey(newPassword, salt)
	hash := us.encoder.EncodeKey(key)

	_, err = us.store.UpdatePassword(ctx, id, hash, salt)
	if err != nil {
		return wrap(err)
	}

	return nil
}

func NewService(store Store, encoder password.Encoder, queue queue.Queue) Service {
	return &userService{
		store:   store,
		encoder: encoder,
		queue:   queue,
	}
}
