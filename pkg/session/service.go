package session

import (
	"context"
	"fmt"
	"identification-service/pkg/client"
	"identification-service/pkg/liberr"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
	"time"
)

const invalidToken = "NA"

type Service interface {
	LoginUser(ctx context.Context, email, password string) (string, string, error)
	LogoutUser(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
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

	refreshToken, err := ss.generator.GenerateRefreshToken()
	if err != nil {
		return wrap(err)
	}

	session, err := NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	if err != nil {
		return wrap(err)
	}

	ctxWt, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	sessionID, err := ss.store.CreateSession(ctxWt, session)
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
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := ss.store.RevokeSession(ctx, refreshToken)
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

	ctxWt, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	session, err := ss.store.GetSession(ctxWt, refreshToken)
	if err != nil {
		return wrap(err)
	}

	err = validateSession(ctxWt, cl.SessionTTL(), session, ss.store, refreshToken)
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

//TODO: THIS FUNCTION IS NOT TESTED, FIND A WAY TO TEST IT
func validateSession(ctx context.Context, sessionTTL int, session Session, store Store, refreshToken string) error {
	if !session.IsExpired(float64(sessionTTL)) {
		return nil
	}

	//TODO: FIX THE LOGIC HERE
	_, err := store.RevokeSession(ctx, refreshToken)
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
