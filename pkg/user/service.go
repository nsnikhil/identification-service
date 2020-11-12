package user

import (
	"database/sql"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/queue"
	"identification-service/pkg/user/internal"
)

//TODO: RENAME (APPEND USER IN THE NAME)
type Service interface {
	CreateUser(name, email, password string) (string, error)
	UpdatePassword(email, oldPassword, newPassword string) error
	GetUserID(email, password string) (string, error)
}

// TODO: RENAME
type userService struct {
	store   internal.Store
	encoder password.Encoder
	queue   queue.Queue
}

func (us *userService) CreateUser(name, email, password string) (string, error) {
	wrap := func(err error) error { return liberr.WithOp("Service.SignUp", err) }

	user, err := internal.NewUser(us.encoder, name, email, password)
	if err != nil {
		return "", wrap(err)
	}

	userID, err := us.store.CreateUser(user)
	if err != nil {
		return "", wrap(err)
	}

	go us.queue.UnsafePush([]byte(userID))

	return userID, nil
}

func (us *userService) GetUserID(email, password string) (string, error) {
	user, err := us.store.GetUser(email)
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("Service.GetUserID"), liberr.InvalidCredentialsError, err)
	}

	err = us.encoder.VerifyPassword(password, user.PasswordHash(), user.PasswordSalt())
	if err != nil {
		return "", liberr.WithArgs(liberr.Operation("Service.GetUserID"), liberr.InvalidCredentialsError, err)
	}

	return user.ID(), nil
}

func (us *userService) UpdatePassword(email, oldPassword, newPassword string) error {
	wrap := func(err error) error { return liberr.WithOp("Service.UpdatePassword", err) }

	err := us.encoder.ValidatePassword(newPassword)
	if err != nil {
		return wrap(err)
	}

	id, err := us.GetUserID(email, oldPassword)
	if err != nil {
		return wrap(err)
	}

	salt, err := us.encoder.GenerateSalt()
	if err != nil {
		return wrap(err)
	}

	key := us.encoder.GenerateKey(newPassword, salt)
	hash := us.encoder.EncodeKey(key)

	_, err = us.store.UpdatePassword(id, hash, salt)
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
