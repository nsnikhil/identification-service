package session_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/session"
	"identification-service/pkg/test"
	"testing"
)

func TestRevokeOldStrategySuccess(t *testing.T) {
	userID := test.NewUUID()
	ctx := context.Background()

	mockSessionStore := &session.MockStore{}
	mockSessionStore.On("RevokeLastNSessions", ctx, userID, 1).Return(int64(1), nil)

	ro := session.NewRevokeOldStrategy(mockSessionStore)

	err := ro.Apply(ctx, userID, 2, 2)
	assert.NoError(t, err)
}

func TestRevokeOldStrategyFailureWhenStoreCallFails(t *testing.T) {
	userID := test.NewUUID()
	ctx := context.Background()

	mockSessionStore := &session.MockStore{}
	mockSessionStore.On("RevokeLastNSessions", ctx, userID, 1).
		Return(int64(0), errors.New("failed to revoke last n sessions"))

	ro := session.NewRevokeOldStrategy(mockSessionStore)

	err := ro.Apply(ctx, userID, 2, 2)
	assert.Error(t, err)
}

func TestRevokeOldStrategyFailureWhenCurrActiveSessionIsLessThanMaxAllowed(t *testing.T) {
	userID := test.NewUUID()
	ctx := context.Background()

	ro := session.NewRevokeOldStrategy(&session.MockStore{})

	err := ro.Apply(ctx, userID, 1, 2)
	assert.Error(t, err)
}
