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
	"identification-service/pkg/test"
	"testing"
)

func TestCreateNewClientSuccess(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(test.ClientPubKey, test.ClientPriKey, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Client")).Return(test.ClientID, nil)

	svc := client.NewService(mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(context.Background(), test.ClientName, test.ClientAccessTokenTTL, test.ClientSessionTTL, test.ClientMaxActiveSessions)
	require.NoError(t, err)
}

func TestCreateNewClientFailureWhenKeyGenerationFails(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, liberr.WithArgs(errors.New("failed to generate key")))

	svc := client.NewService(&client.MockStore{}, mockKeyGenerator)

	_, _, err := svc.CreateClient(context.Background(), test.ClientName, test.ClientAccessTokenTTL, test.ClientSessionTTL, test.ClientMaxActiveSessions)
	require.Error(t, err)
}

func TestCreateNewClientFailureWhenStoreReturnFailure(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(test.ClientPubKey, test.ClientPriKey, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("Client")).Return("", liberr.WithArgs(errors.New("failed to create client")))

	svc := client.NewService(mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(context.Background(), test.ClientName, test.ClientAccessTokenTTL, test.ClientSessionTTL, test.ClientMaxActiveSessions)
	require.Error(t, err)
}

func TestRevokeClientSuccess(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.timerCtx"), test.ClientID).Return(int64(1), nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), test.ClientID)
	require.NoError(t, err)
}

func TestRevokeClientFailure(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.timerCtx"), test.ClientID).Return(int64(0), liberr.WithArgs(errors.New("failed to revoke client")))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), test.ClientID)
	require.Error(t, err)
}

func TestGetClientSuccess(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), test.ClientName, test.ClientSecret).Return(client.Client{}, nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	_, err := svc.GetClient(context.Background(), test.ClientName, test.ClientSecret)
	require.NoError(t, err)
}

func TestGetClientFailure(t *testing.T) {
	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.timerCtx"), test.ClientName, test.ClientSecret).Return(client.Client{}, liberr.WithArgs(errors.New("failed to get client")))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	_, err := svc.GetClient(context.Background(), test.ClientName, test.ClientSecret)
	require.Error(t, err)
}
