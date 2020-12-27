package client_test

import (
	"context"
	"crypto/ed25519"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/test"
	"testing"
)

func TestCreateNewClientSuccess(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(test.ClientPubKey, test.ClientPriKey, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("client.Client")).Return(test.ClientID(), nil)

	svc := client.NewService(mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		test.ClientName(),
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
	)

	require.NoError(t, err)
}

func TestCreateNewClientFailureWhenClientValidationFails(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(test.ClientPubKey, test.ClientPriKey, nil)

	svc := client.NewService(&client.MockStore{}, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		"",
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
	)

	require.Error(t, err)
}

func TestCreateNewClientFailureWhenKeyGenerationFails(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, errors.New("failed to generate key"))

	svc := client.NewService(&client.MockStore{}, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		test.ClientName(),
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
	)

	require.Error(t, err)
}

func TestCreateNewClientFailureWhenStoreReturnFailure(t *testing.T) {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(test.ClientPubKey, test.ClientPriKey, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("client.Client")).Return("", errors.New("failed to create client"))

	svc := client.NewService(mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		test.ClientName(),
		test.ClientAccessTokenTTL,
		test.ClientSessionTTL,
		test.ClientMaxActiveSessions,
		test.ClientSessionStrategyRevokeOld,
	)

	require.Error(t, err)
}

func TestRevokeClientSuccess(t *testing.T) {
	clientID := test.ClientID()

	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.emptyCtx"), clientID).Return(int64(1), nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), clientID)
	require.NoError(t, err)
}

func TestRevokeClientFailure(t *testing.T) {
	clientID := test.ClientID()

	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.emptyCtx"), clientID).Return(int64(0), errors.New("failed to revoke client"))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), clientID)
	require.Error(t, err)
}

func TestGetClientSuccess(t *testing.T) {
	clientName, clientSecret := test.ClientName(), test.ClientSecret()

	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(client.Client{}, nil)

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	_, err := svc.GetClient(context.Background(), clientName, clientSecret)
	require.NoError(t, err)
}

func TestGetClientFailure(t *testing.T) {
	clientName, clientSecret := test.ClientName(), test.ClientSecret()

	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(client.Client{}, errors.New("failed to get client"))

	svc := client.NewService(mockStore, &libcrypto.MockEd25519Generator{})

	_, err := svc.GetClient(context.Background(), clientName, clientSecret)
	require.Error(t, err)
}
