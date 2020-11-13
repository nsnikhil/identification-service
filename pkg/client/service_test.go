package client_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/client/internal"
	"identification-service/pkg/liberr"
	"testing"
)

const (
	name           = "clientOne"
	secret         = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
	accessTokenTTL = 10
	sessionTTL     = 14400
	id             = "f14abb31-ec1a-4ff6-a937-c2e930ca34ef"
)

func TestCreateNewClientSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.Client")).Return(id, nil)

	svc := client.NewInternalService(mockStore)

	_, err := svc.CreateClient(context.Background(), name, accessTokenTTL, sessionTTL)
	require.NoError(t, err)
}

func TestCreateNewClientFailure(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("internal.Client")).Return("", liberr.WithArgs(errors.New("failed to create client")))

	svc := client.NewInternalService(mockStore)

	_, err := svc.CreateClient(context.Background(), name, accessTokenTTL, sessionTTL)
	require.Error(t, err)
}

func TestRevokeClientSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.timerCtx"), id).Return(int64(1), nil)

	svc := client.NewInternalService(mockStore)

	err := svc.RevokeClient(context.Background(), id)
	require.NoError(t, err)
}

func TestRevokeClientFailure(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.timerCtx"), id).Return(int64(0), liberr.WithArgs(errors.New("failed to revoke client")))

	svc := client.NewInternalService(mockStore)

	err := svc.RevokeClient(context.Background(), id)
	require.Error(t, err)
}

func TestGetClientTTLSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(internal.Client{}, nil)

	svc := client.NewInternalService(mockStore)

	_, _, err := svc.GetClientTTL(context.Background(), name, secret)
	require.NoError(t, err)
}

func TestGetClientTTLFailure(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(internal.Client{}, liberr.WithArgs(errors.New("failed to get client")))

	svc := client.NewInternalService(mockStore)

	_, _, err := svc.GetClientTTL(context.Background(), name, secret)
	require.Error(t, err)
}

func TestValidateClientSuccess(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(internal.Client{}, nil)

	svc := client.NewInternalService(mockStore)

	err := svc.ValidateClientCredentials(context.Background(), name, secret)
	require.NoError(t, err)
}

func TestValidateClientFailure(t *testing.T) {
	mockStore := &internal.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(internal.Client{}, liberr.WithArgs(errors.New("failed to get client")))

	svc := client.NewInternalService(mockStore)

	err := svc.ValidateClientCredentials(context.Background(), name, secret)
	require.Error(t, err)
}
