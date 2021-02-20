package consumer_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"identification-service/pkg/consumer"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
)

func TestUpdatePasswordHandlerSuccess(t *testing.T) {
	userID := test.NewUUID()

	mockSessionService := &session.MockService{}
	mockSessionService.On("RevokeAllSessions", mock.AnythingOfType("*context.emptyCtx"), userID).
		Return(nil)

	uph := consumer.NewUpdatePasswordHandler(mockSessionService)

	err := uph.Handle([]byte(userID))
	assert.NoError(t, err)
}

func TestUpdatePasswordHandlerFailure(t *testing.T) {
	userID := test.NewUUID()

	mockSessionService := &session.MockService{}
	mockSessionService.On("RevokeAllSessions", mock.AnythingOfType("*context.emptyCtx"), userID).
		Return(errors.New("failed to revoke all sessions"))

	uph := consumer.NewUpdatePasswordHandler(mockSessionService)

	err := uph.Handle([]byte(userID))
	assert.Error(t, err)
}
