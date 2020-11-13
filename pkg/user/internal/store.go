package internal

import (
	"context"
	"database/sql"
	"fmt"
	"identification-service/pkg/liberr"
)

const (
	insertUser     = `insert into users (name, email, passwordhash, passwordsalt) values ($1, $2, $3, $4) returning id`
	getUserByEmail = `select id, name, email, passwordhash, passwordsalt from users where email = $1`
	updatePassword = `update users set passwordhash=$1, passwordsalt=$2 where id=$3`
)

type Store interface {
	CreateUser(ctx context.Context, user User) (string, error)
	GetUser(ctx context.Context, email string) (User, error)
	UpdatePassword(ctx context.Context, userID string, newPasswordHash string, newPasswordSalt []byte) (int64, error)
}

// TODO: RENAME
type userStore struct {
	db *sql.DB
}

func (us *userStore) CreateUser(ctx context.Context, user User) (string, error) {
	var id string

	//TODO: RETURN DIFFERENT ERROR KIND FOR DUPLICATE RECORD
	err := us.db.QueryRow(insertUser, user.name, user.email, user.passwordHash, user.passwordSalt).Scan(&id)
	if err != nil {
		return "", liberr.WithOp("Store.CreateUser", err)
	}

	return id, nil
}

func (us *userStore) GetUser(ctx context.Context, email string) (User, error) {
	var user User

	row := us.db.QueryRowContext(context.Background(), getUserByEmail, email)
	if row.Err() != nil {
		return user, liberr.WithOp("Store.GetUser", row.Err())
	}

	err := row.Scan(&user.id, &user.name, &user.email, &user.passwordHash, &user.passwordSalt)
	if err != nil {
		return user, liberr.WithOp("Store.GetUser", err)
	}

	return user, nil
}

func (us *userStore) UpdatePassword(ctx context.Context, userID string, newPasswordHash string, newPasswordSalt []byte) (int64, error) {
	wrap := func(err error) error { return liberr.WithOp("Store.UpdatePassword", err) }

	res, err := us.db.ExecContext(context.Background(), updatePassword, newPasswordHash, newPasswordSalt, userID)
	if err != nil {
		return 0, wrap(err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return 0, wrap(err)
	}

	if c == 0 {
		return 0, fmt.Errorf("no record found with id %s", userID)
	}

	return c, nil
}

func NewStore(db *sql.DB) Store {
	return &userStore{
		db: db,
	}
}
