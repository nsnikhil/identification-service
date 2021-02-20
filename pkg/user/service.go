package user

import (
	"context"
	"identification-service/pkg/config"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/producer"
)

//TODO: RENAME (APPEND USER IN THE NAME)
type Service interface {
	CreateUser(ctx context.Context, name, email, password string) (string, error)
	UpdatePassword(ctx context.Context, email, oldPassword, newPassword string) error
	GetUserID(ctx context.Context, email, password string) (string, error)
}

// TODO: RENAME
type userService struct {
	cfg     config.KafkaConfig
	store   Store
	encoder password.Encoder
	pdr     producer.Producer
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

	//TODO: CHECK FOR ERROR
	go us.pdr.Produce(us.cfg.SignUpTopicName(), []byte(userID))

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

	userID, err := us.GetUserID(ctx, email, oldPassword)
	if err != nil {
		return wrap(err)
	}

	salt, err := us.encoder.GenerateSalt()
	if err != nil {
		return wrap(err)
	}

	key := us.encoder.GenerateKey(newPassword, salt)
	hash := us.encoder.EncodeKey(key)

	_, err = us.store.UpdatePassword(ctx, userID, hash, salt)
	if err != nil {
		return wrap(err)
	}

	//TODO: CHECK FOR ERROR
	go us.pdr.Produce(us.cfg.UpdatePasswordTopicName(), []byte(userID))

	return nil
}

func NewService(cfg config.KafkaConfig, store Store, encoder password.Encoder, pdr producer.Producer) Service {
	return &userService{
		cfg:     cfg,
		store:   store,
		encoder: encoder,
		pdr:     pdr,
	}
}
