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
	strategies  map[string]Strategy
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
		strategy, ok := ss.strategies[cl.SessionStrategyName()]
		if !ok {
			return wrap(fmt.Errorf("invalid sesion strategy %s", cl.SessionStrategyName()))
		}

		err = strategy.Apply(ctx, userID, activeSessionsCount, cl.MaxActiveSessions())
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
	wrap := func(err error) error {
		return liberr.WithOp("Service.LogoutUser", err)
	}

	cl, err := client.FromContext(ctx)
	if err != nil {
		return wrap(err)
	}

	_, err = getValidSession(ctx, cl, ss.store, refreshToken)
	if err != nil {
		return wrap(err)
	}

	_, err = ss.store.RevokeSessions(ctx, refreshToken)
	if err != nil {
		return wrap(err)
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

	session, err := getValidSession(ctx, cl, ss.store, refreshToken)
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

func getValidSession(ctx context.Context, cl client.Client, store Store, refreshToken string) (Session, error) {
	session, err := store.GetSession(ctx, refreshToken)
	if err != nil {
		return Session{}, err
	}

	err = validateSession(cl.SessionTTL(), session, refreshToken)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func validateSession(sessionTTL int, session Session, refreshToken string) error {
	if session.revoked || session.IsExpired(float64(sessionTTL)) {
		return liberr.WithArgs(liberr.AuthenticationError, fmt.Errorf("session expired for %s", refreshToken))
	}

	return nil
}

func NewService(store Store, userService user.Service, generator token.Generator, strategies map[string]Strategy) Service {
	return &sessionService{
		store:       store,
		userService: userService,
		generator:   generator,
		strategies:  strategies,
	}
}
