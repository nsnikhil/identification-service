package session

import (
	"context"
	"fmt"
	"identification-service/pkg/client"
	"identification-service/pkg/liberr"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
)

const invalidToken = "NA"

type Service interface {
	LoginUser(ctx context.Context, email, password string) (string, string, error)
	LogoutUser(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	RevokeAllSessions(ctx context.Context, userID string) error
}

type sessionService struct {
	store       Store
	userService user.Service
	generator   token.Generator
}

func (ss *sessionService) LoginUser(ctx context.Context, email, password string) (string, string, error) {
	wrap := func(err error) (string, string, error) {
		return invalidToken, invalidToken, liberr.WithOp("Service.LoginUser", err)
	}

	cl, err := client.FromContext(ctx)
	if err != nil {
		return wrap(err)
	}

	userID, err := ss.userService.GetUserID(ctx, email, password)
	if err != nil {
		return wrap(err)
	}

	activeSessionsCount, err := ss.store.GetActiveSessionsCount(ctx, userID)
	if err != nil {
		return wrap(err)
	}

	if activeSessionsCount >= cl.MaxActiveSessions() {
		strategy, err := strategyFromName(cl.SessionStrategyName(), ss.store)
		if err != nil {
			return wrap(err)
		}

		err = strategy.apply(ctx, userID, activeSessionsCount, cl.MaxActiveSessions())
		if err != nil {
			return wrap(err)
		}
	}

	refreshToken, err := ss.generator.GenerateRefreshToken()
	if err != nil {
		return wrap(err)
	}

	session, err := NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	if err != nil {
		return wrap(err)
	}

	sessionID, err := ss.store.CreateSession(ctx, session)
	if err != nil {
		return wrap(err)
	}

	accessToken, err := ss.generator.GenerateAccessToken(
		cl.AccessTokenTTL(),
		userID, map[string]string{"session_id": sessionID},
	)

	if err != nil {
		return wrap(err)
	}

	return accessToken, refreshToken, nil
}

func (ss *sessionService) LogoutUser(ctx context.Context, refreshToken string) error {
	_, err := ss.store.RevokeSessions(ctx, refreshToken)
	if err != nil {
		return liberr.WithOp("Service.LogoutUser", err)
	}

	return nil
}

func (ss *sessionService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	wrap := func(err error) (string, error) {
		return invalidToken, liberr.WithOp("Service.RefreshToken", err)
	}

	cl, err := client.FromContext(ctx)
	if err != nil {
		return wrap(err)
	}

	session, err := ss.store.GetSession(ctx, refreshToken)
	if err != nil {
		return wrap(err)
	}

	err = validateSession(ctx, cl.SessionTTL(), session, ss.store, refreshToken)
	if err != nil {
		return wrap(err)
	}

	accessToken, err := ss.generator.GenerateAccessToken(
		cl.AccessTokenTTL(),
		session.userID,
		map[string]string{"session_id": session.id},
	)

	if err != nil {
		return wrap(err)
	}

	return accessToken, nil
}

func (ss *sessionService) RevokeAllSessions(ctx context.Context, userID string) error {
	_, err := ss.store.RevokeAllSessions(ctx, userID)
	if err != nil {
		return liberr.WithOp("Service.RevokeAllSessions", err)
	}

	return nil
}

func validateSession(ctx context.Context, sessionTTL int, session Session, store Store, refreshToken string) error {
	if !session.IsExpired(float64(sessionTTL)) {
		return nil
	}

	//TODO: FIX THE LOGIC HERE
	_, err := store.RevokeSessions(ctx, refreshToken)
	if err != nil {
		return err
	}

	return fmt.Errorf("session expired for %s", refreshToken)
}

func NewService(store Store, userService user.Service, generator token.Generator) Service {
	return &sessionService{
		store:       store,
		userService: userService,
		generator:   generator,
	}
}
