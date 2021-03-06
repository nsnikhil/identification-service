package user

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"github.com/nsnikhil/erx"
	"identification-service/pkg/database"
)

const (
	insertUser     = `insert into users (name, email, password_hash, password_salt) values ($1, $2, $3, $4) returning id`
	getUserByEmail = `select id, name, email, password_hash, password_salt from users where email = $1`
	updatePassword = `update users set password_hash=$1, password_salt=$2 where id=$3`
)

type Store interface {
	CreateUser(ctx context.Context, user User) (string, error)
	GetUser(ctx context.Context, email string) (User, error)
	UpdatePassword(ctx context.Context, userID string, newPasswordHash string, newPasswordSalt []byte) (int64, error)
}

// TODO: RENAME
type userStore struct {
	db database.SQLDatabase
}

func (us *userStore) CreateUser(ctx context.Context, user User) (string, error) {
	var id string

	//TODO: REMOVE THIS HARD CODING
	row := us.db.QueryRowContext(ctx, insertUser, user.name, user.email, user.passwordHash, user.passwordSalt)
	if row.Err() != nil {
		if pgErr, ok := row.Err().(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return "", erx.WithArgs(erx.Operation("Store.CreateUser"), erx.DuplicateRecordError, row.Err())
			}
		}

		return "", erx.WithArgs(erx.Operation("Store.CreateUser"), row.Err())
	}

	err := row.Scan(&id)
	if err != nil {
		return "", erx.WithArgs(erx.Operation("Store.CreateUser"), err)
	}

	return id, nil
}

func (us *userStore) GetUser(ctx context.Context, email string) (User, error) {
	var user User

	row := us.db.QueryRowContext(context.Background(), getUserByEmail, email)
	if row.Err() != nil {
		return user, erx.WithArgs(erx.Operation("Store.GetUser"), row.Err())
	}

	err := row.Scan(&user.id, &user.name, &user.email, &user.passwordHash, &user.passwordSalt)
	if err != nil {
		return user, erx.WithArgs(erx.Operation("Store.GetUser"), err)
	}

	return user, nil
}

func (us *userStore) UpdatePassword(ctx context.Context, userID string, newPasswordHash string, newPasswordSalt []byte) (int64, error) {
	wrap := func(err error) error { return erx.WithArgs(erx.Operation("Store.UpdatePassword"), err) }

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

func NewStore(db database.SQLDatabase) Store {
	return &userStore{
		db: db,
	}
}
