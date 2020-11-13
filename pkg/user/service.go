package user

import (
	"context"
	"database/sql"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/queue"
	"identification-service/pkg/user/internal"
	"time"
)

//TODO: RENAME (APPEND USER IN THE NAME)
type Service interface {
	CreateUser(ctx context.Context, name, email, password string) (string, error)
	UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error
	GetUserID(ctx context.Context, email, password string) (string, error)
}

// TODO: RENAME
type userService struct {
	store   internal.Store
	encoder password.Encoder
	queue   queue.Queue
}

func (us *userService) CreateUser(ctx context.Context, name, email, password string) (string, error) {
	wrap := func(err error) error { return liberr.WithOp("Service.SignUp", err) }

	user, err := internal.NewUser(us.encoder, name, email, password)
	if err != nil {
		return "", wrap(err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	userID, err := us.store.CreateUser(ctx, user)
	if err != nil {
		return "", wrap(err)
	}

	go us.queue.UnsafePush([]byte(userID))

	return userID, nil
}

func (us *userService) GetUserID(ctx context.Context, email, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	user, err := us.store.GetUser(ctx, email)
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("Service.GetUserID"), liberr.InvalidCredentialsError, err)
	}

	err = us.encoder.VerifyPassword(password, user.PasswordHash(), user.PasswordSalt())
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("Service.GetUserID"), liberr.InvalidCredentialsError, err)
	}

	return user.ID(), nil
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

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err = us.store.UpdatePassword(ctx, id, hash, salt)
	if err != nil {
		return wrap(err)
	}

	return nil
}

//TODO: ONLY USED IN TESTS
func NewInternalService(store internal.Store, encoder password.Encoder, queue queue.Queue) Service {
	return &userService{
		store:   store,
		encoder: encoder,
		queue:   queue,
	}
}

func NewService(db *sql.DB, encoder password.Encoder, queue queue.Queue) Service {
	return &userService{
		store:   internal.NewStore(db),
		encoder: encoder,
		queue:   queue,
	}
}
