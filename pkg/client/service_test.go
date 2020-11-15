package client_test

import (
	"context"
	"crypto/ed25519"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/liberr"
	"testing"
)

func TestCreateNewClientSuccess(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(pubKey, priKey, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Client")).Return(id, nil)

	svc := client.NewService(mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(context.Background(), name, accessTokenTTL, sessionTTL, maxActiveSessions)
	require.NoError(t, err)
}

func TestCreateNewClientFailureWhenKeyGenerationFails(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, liberr.WithArgs(errors.New("failed to generate key")))

	svc := client.NewService(&client.MockStore{}, mockKeyGenerator)

	_, _, err := svc.CreateClient(context.Background(), name, accessTokenTTL, sessionTTL, maxActiveSessions)
	require.Error(t, err)
}

func TestCreateNewClientFailureWhenStoreReturnFailure(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(pubKey, priKey, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Client")).Return("", liberr.WithArgs(errors.New("failed to create client")))

	svc := client.NewService(mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(context.Background(), name, accessTokenTTL, sessionTTL, maxActiveSessions)
	require.Error(t, err)
}

func TestRevokeClientSuccess(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.timerCtx"), id).Return(int64(1), nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), id)
	require.NoError(t, err)
}

func TestRevokeClientFailure(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.timerCtx"), id).Return(int64(0), liberr.WithArgs(errors.New("failed to revoke client")))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), id)
	require.Error(t, err)
}

func TestGetClientTTLSuccess(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(client.Client{}, nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	_, _, err := svc.GetClientTTL(context.Background(), name, secret)
	require.NoError(t, err)
}

func TestGetClientTTLFailure(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(client.Client{}, liberr.WithArgs(errors.New("failed to get client")))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	_, _, err := svc.GetClientTTL(context.Background(), name, secret)
	require.Error(t, err)
}

func TestValidateClientSuccess(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(client.Client{}, nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.ValidateClientCredentials(context.Background(), name, secret)
	require.NoError(t, err)
}

func TestValidateClientFailure(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), name, secret).Return(client.Client{}, liberr.WithArgs(errors.New("failed to get client")))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.ValidateClientCredentials(context.Background(), name, secret)
	require.Error(t, err)
}
