package session

import (
	"context"
	"database/sql"
	"fmt"
	"identification-service/pkg/liberr"
)

const (
	createSession = `insert into sessions (user_id, refresh_token) values ($1, $2) returning id`
	getSession    = `select id, user_id, revoked, created_at, updated_at from sessions where refresh_token=$1`
	revokeSession = `update sessions set revoked=true where refresh_token=$1`
)

type Store interface {
	CreateSession(ctx context.Context, session Session) (string, error)
	GetSession(ctx context.Context, refreshToken string) (Session, error)
	RevokeSession(ctx context.Context, refreshToken string) (int64, error)
}

type sessionStore struct {
	db *sql.DB
}

func (ss *sessionStore) CreateSession(ctx context.Context, session Session) (string, error) {
	var sessionID string

	err := ss.db.QueryRowContext(ctx, createSession, session.userID, session.refreshToken).Scan(&sessionID)
	if err != nil {
		return "", liberr.WithOp("Store.CreateSession", err)
	}

	return sessionID, nil
}

func (ss *sessionStore) GetSession(ctx context.Context, refreshToken string) (Session, error) {
	var session Session

	row := ss.db.QueryRowContext(ctx, getSession, refreshToken)
	if row.Err() != nil {
		return session, liberr.WithOp("Store.GetSession", row.Err())
	}

	err := row.Scan(&session.id, &session.userID, &session.revoked, &session.createdAt, &session.updatedAt)
	if err != nil {
		return session, liberr.WithOp("Store.GetSession", err)
	}

	return session, nil
}

func (ss *sessionStore) RevokeSession(ctx context.Context, refreshToken string) (int64, error) {
	res, err := ss.db.ExecContext(ctx, revokeSession, refreshToken)
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeSession", err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeSession", err)
	}

	if c == 0 {
		return 0, liberr.WithOp(
			"Store.RevokeSession",
			fmt.Errorf("no session found for refresh token %s", refreshToken),
		)
	}

	return c, nil
}

func NewStore(db *sql.DB) Store {
	return &sessionStore{
		db: db,
	}
}
