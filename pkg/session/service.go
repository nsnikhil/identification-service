package session

import (
	"database/sql"
	"fmt"
	"identification-service/pkg/client"
	"identification-service/pkg/liberr"
	"identification-service/pkg/session/internal"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
)

const invalidToken = "NA"

//TODO: SHOULD CLIENT NAME AND SECRET BE PART OF CONTEXT ?
type Service interface {
	LoginUser(clientName, clientSecret, email, password string) (string, string, error)
	LogoutUser(refreshToken string) error
	RefreshToken(clientName, clientSecret, refreshToken string) (string, error)
}

type sessionService struct {
	store         internal.Store
	userService   user.Service
	clientService client.Service
	generator     token.Generator
}

func (ss *sessionService) LoginUser(clientName, clientSecret, email, password string) (string, string, error) {
	wrap := func(err error) (string, string, error) {
		return invalidToken, invalidToken, liberr.WithOp("Service.LoginUser", err)
	}

	userID, err := ss.userService.GetUserID(email, password)
	if err != nil {
		return wrap(err)
	}

	refreshToken, err := ss.generator.GenerateRefreshToken()
	if err != nil {
		return wrap(err)
	}

	session, err := internal.NewSessionBuilder().UserID(userID).RefreshToken(refreshToken).Build()
	if err != nil {
		return wrap(err)
	}

	sessionID, err := ss.store.CreateSession(session)
	if err != nil {
		return wrap(err)
	}

	accessTokenTTL, _, err := ss.clientService.GetClientTTL(clientName, clientSecret)
	if err != nil {
		return wrap(err)
	}

	accessToken, err := ss.generator.GenerateAccessToken(accessTokenTTL, userID, map[string]string{"session_id": sessionID})
	if err != nil {
		return wrap(err)
	}

	return accessToken, refreshToken, nil
}

func (ss *sessionService) LogoutUser(refreshToken string) error {
	_, err := ss.store.RevokeSession(refreshToken)
	if err != nil {
		return liberr.WithOp("Service.LogoutUser", err)
	}

	return nil
}

func (ss *sessionService) RefreshToken(clientName, clientSecret, refreshToken string) (string, error) {
	wrap := func(err error) (string, error) {
		return invalidToken, liberr.WithOp("Service.RefreshToken", err)
	}

	session, err := ss.store.GetSession(refreshToken)
	if err != nil {
		return wrap(err)
	}

	accessTokenTTL, sessionTTL, err := ss.clientService.GetClientTTL(clientName, clientSecret)
	if err != nil {
		return wrap(err)
	}

	err = validateSession(sessionTTL, session, ss.store, refreshToken)
	if err != nil {
		return wrap(err)
	}

	accessToken, err := ss.generator.GenerateAccessToken(accessTokenTTL, session.UserID(), map[string]string{"session_id": session.ID()})
	if err != nil {
		return wrap(err)
	}

	return accessToken, nil
}

//TODO: THIS FUNCTION IS NOT TESTED, FIND A WAY TO TEST IT
func validateSession(sessionTTL int, session internal.Session, store internal.Store, refreshToken string) error {
	if !session.IsExpired(float64(sessionTTL)) {
		return nil
	}

	//TODO: FIX THE LOGIC HERE
	_, err := store.RevokeSession(refreshToken)
	if err != nil {
		return err
	}

	return fmt.Errorf("session expired for %s", refreshToken)
}

func NewInternalService(store internal.Store, userService user.Service, clientService client.Service, generator token.Generator) Service {
	return &sessionService{
		store:         store,
		userService:   userService,
		clientService: clientService,
		generator:     generator,
	}
}

func NewService(db *sql.DB, userService user.Service, clientService client.Service, generator token.Generator) Service {
	return &sessionService{
		store:         internal.NewStore(db),
		userService:   userService,
		clientService: clientService,
		generator:     generator,
	}
}
