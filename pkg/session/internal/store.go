package internal

import (
	"database/sql"
	"fmt"
	"identification-service/pkg/liberr"
)

const (
	createSession = `insert into sessions (userid, refreshtoken) values ($1, $2) returning id`
	getSession    = `select id, userid, revoked, createdat, updatedat from sessions where refreshtoken=$1`
	revokeSession = `update sessions set revoked=true where refreshtoken=$1`
)

type Store interface {
	CreateSession(session Session) (string, error)
	GetSession(refreshToken string) (Session, error)
	RevokeSession(refreshToken string) (int64, error)
}

type sessionStore struct {
	db *sql.DB
}

func (ss *sessionStore) CreateSession(session Session) (string, error) {
	var sessionID string

	err := ss.db.QueryRow(createSession, session.userID, session.refreshToken).Scan(&sessionID)
	if err != nil {
		return "", liberr.WithOp("Store.CreateSession", err)
	}

	return sessionID, nil
}

func (ss *sessionStore) GetSession(refreshToken string) (Session, error) {
	var session Session

	row := ss.db.QueryRow(getSession, refreshToken)
	if row.Err() != nil {
		return session, liberr.WithOp("Store.GetSession", row.Err())
	}

	err := row.Scan(&session.id, &session.userID, &session.revoked, &session.createdAt, &session.updatedAt)
	if err != nil {
		return session, liberr.WithOp("Store.GetSession", err)
	}

	return session, nil
}

func (ss *sessionStore) RevokeSession(refreshToken string) (int64, error) {
	res, err := ss.db.Exec(revokeSession, refreshToken)
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeSession", err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeSession", err)
	}

	if c == 0 {
		return 0, liberr.WithOp("Store.RevokeSession", fmt.Errorf("no seesion found for refresh token %s", refreshToken))
	}

	return c, nil
}

func NewStore(db *sql.DB) Store {
	return &sessionStore{
		db: db,
	}
}
