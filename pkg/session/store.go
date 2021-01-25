package session

import (
	"context"
	"fmt"
	"identification-service/pkg/database"
	"identification-service/pkg/liberr"
	"strings"
)

const (
	createSession          = `insert into sessions (user_id, refresh_token) values ($1, $2) returning id`
	getSession             = `select id, user_id, revoked, created_at, updated_at from sessions where refresh_token=$1`
	getActiveSessionsCount = `select count(*) from sessions where user_id=$1 and revoked=false`
	revokeSessions         = `update sessions set revoked=true where refresh_token = ANY($1::uuid[])`
	getLastNRefreshTokens  = `select refresh_token from sessions where user_id=$1 and revoked=false order by created_at asc limit $2`
	revokeAllSessions      = `update sessions set revoked=true where user_id=$1`
)

type Store interface {
	CreateSession(ctx context.Context, session Session) (string, error)
	GetSession(ctx context.Context, refreshToken string) (Session, error)
	GetActiveSessionsCount(ctx context.Context, userID string) (int, error)
	RevokeSessions(ctx context.Context, refreshTokens ...string) (int64, error)

	RevokeAllSessions(ctx context.Context, userID string) (int64, error)

	//TODO: REFACTOR
	RevokeLastNSessions(ctx context.Context, userID string, n int) (int64, error)
}

type sessionStore struct {
	db database.SQLDatabase
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

func (ss *sessionStore) GetActiveSessionsCount(ctx context.Context, userID string) (int, error) {
	var activeSessionCount int

	err := ss.db.QueryRowContext(ctx, getActiveSessionsCount, userID).Scan(&activeSessionCount)
	if err != nil {
		return -1, liberr.WithOp("Store.GetActiveSessionsCount", err)
	}

	return activeSessionCount, nil
}

func (ss *sessionStore) RevokeSessions(ctx context.Context, refreshTokens ...string) (int64, error) {
	res, err := ss.db.ExecContext(ctx, revokeSessions, toArgs(refreshTokens))
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeSession", err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeSession", err)
	}

	if c == 0 {
		return 0, liberr.WithArgs(
			liberr.Operation("Store.RevokeSession"),
			liberr.ResourceNotFound,
			fmt.Errorf("no session found for refresh tokens %v", refreshTokens),
		)
	}

	return c, nil
}

func (ss *sessionStore) RevokeLastNSessions(ctx context.Context, userID string, n int) (int64, error) {
	rows, err := ss.db.QueryContext(ctx, getLastNRefreshTokens, userID, n)
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeLastNSessions", err)
	}

	var refreshTokens []string

	for rows.Next() {
		var refreshToken string

		err := rows.Scan(&refreshToken)
		if err != nil {
			return 0, liberr.WithOp("Store.RevokeLastNSessions", err)
		}

		refreshTokens = append(refreshTokens, refreshToken)
	}

	if len(refreshTokens) == 0 {
		return 0, liberr.WithOp(
			"Store.RevokeLastNSessions",
			fmt.Errorf("no refresh tokens found to revoke against %s", userID),
		)
	}

	return ss.RevokeSessions(ctx, refreshTokens...)
}

func (ss *sessionStore) RevokeAllSessions(ctx context.Context, userID string) (int64, error) {
	res, err := ss.db.ExecContext(ctx, revokeAllSessions, userID)
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeAllSessions", err)
	}

	c, err := res.RowsAffected()
	if err != nil {
		return 0, liberr.WithOp("Store.RevokeAllSessions", err)
	}

	if c == 0 {
		return 0, liberr.WithOp(
			"Store.RevokeAllSessions",
			fmt.Errorf("no sessions found for user %s", userID),
		)
	}

	return c, nil
}

func toArgs(values []string) string {
	return "{" + strings.Join(values, ",") + "}"
}

func NewStore(db database.SQLDatabase) Store {
	return &sessionStore{
		db: db,
	}
}
