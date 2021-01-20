package client_test

import (
	"context"
	"crypto/ed25519"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/test"
	"testing"
)

type clientServiceTest struct {
	suite.Suite
	cfg config.ClientConfig
}

func (cst *clientServiceTest) SetupSuite() {
	mockClientConfig := &config.MockClientConfig{}
	mockClientConfig.On("Strategies").
		Return(map[string]bool{test.ClientSessionStrategyRevokeOld: true})

	cst.cfg = mockClientConfig
}

func TestClientService(t *testing.T) {
	suite.Run(t, new(clientServiceTest))
}

func (cst *clientServiceTest) TestCreateNewClientSuccess() {
	pub, pri := test.GenerateKey()

	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(pub, pri, nil)

	mockStore := &client.MockStore{}
	mockStore.On(
		"CreateClient",
		mock.AnythingOfType("*context.emptyCtx"),
		mock.AnythingOfType("client.Client"),
	).Return(test.NewUUID(), nil)

	mockClientConfig := &config.MockClientConfig{}
	mockClientConfig.On("Strategies").
		Return(map[string]bool{test.ClientSessionStrategyRevokeOld: true})

	svc := client.NewService(mockClientConfig, mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		test.RandString(8),
		test.RandInt(1, 10),
		test.RandInt(1440, 86701),
		test.RandInt(1, 10),
		test.ClientSessionStrategyRevokeOld,
	)

	cst.Require().NoError(err)
}

func (cst *clientServiceTest) TestCreateNewClientFailureWhenClientValidationFails() {
	pub, pri := test.GenerateKey()

	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(pub, pri, nil)

	svc := client.NewService(cst.cfg, &client.MockStore{}, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		"",
		test.RandInt(1, 10),
		test.RandInt(1440, 86701),
		test.RandInt(1, 10),
		test.ClientSessionStrategyRevokeOld,
	)

	cst.Require().Error(err)
}

func (cst *clientServiceTest) TestCreateNewClientFailureWhenKeyGenerationFails() {
	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(ed25519.PublicKey{}, ed25519.PrivateKey{}, errors.New("failed to generate key"))

	svc := client.NewService(cst.cfg, &client.MockStore{}, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		test.RandString(8),
		test.RandInt(1, 10),
		test.RandInt(1440, 86701),
		test.RandInt(1, 10),
		test.ClientSessionStrategyRevokeOld,
	)

	cst.Require().Error(err)
}

func (cst *clientServiceTest) TestCreateNewClientFailureWhenStoreReturnFailure() {
	pub, pri := test.GenerateKey()

	mockKeyGenerator := &libcrypto.MockEd25519Generator{}
	mockKeyGenerator.On("Generate").Return(pub, pri, nil)

	mockStore := &client.MockStore{}
	mockStore.On("CreateClient", mock.AnythingOfType("*context.emptyCtx"), mock.AnythingOfType("client.Client")).Return("", errors.New("failed to create client"))

	svc := client.NewService(cst.cfg, mockStore, mockKeyGenerator)

	_, _, err := svc.CreateClient(
		context.Background(),
		test.RandString(8),
		test.RandInt(1, 10),
		test.RandInt(1440, 86701),
		test.RandInt(1, 10),
		test.ClientSessionStrategyRevokeOld,
	)

	cst.Require().Error(err)
}

func (cst *clientServiceTest) TestRevokeClientSuccess() {
	clientID := test.NewUUID()

	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.emptyCtx"), clientID).Return(int64(1), nil)

	svc := client.NewService(cst.cfg, mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), clientID)
	cst.Require().NoError(err)
}

func (cst *clientServiceTest) TestRevokeClientFailure() {
	clientID := test.NewUUID()

	mockStore := &client.MockStore{}
	mockStore.On("RevokeClient", mock.AnythingOfType("*context.emptyCtx"), clientID).Return(int64(0), errors.New("failed to revoke client"))

	svc := client.NewService(cst.cfg, mockStore, &libcrypto.MockEd25519Generator{})

	err := svc.RevokeClient(context.Background(), clientID)
	cst.Require().Error(err)
}

func (cst *clientServiceTest) TestGetClientSuccess() {
	clientName, clientSecret := test.RandString(8), test.NewUUID()

	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(client.Client{}, nil)

	svc := client.NewService(cst.cfg, mockStore, &libcrypto.MockEd25519Generator{})

	_, err := svc.GetClient(context.Background(), clientName, clientSecret)
	cst.Require().NoError(err)
}

func (cst *clientServiceTest) TestGetClientFailure() {
	clientName, clientSecret := test.RandString(8), test.NewUUID()

	mockStore := &client.MockStore{}
	mockStore.On("GetClient", mock.AnythingOfType("*context.emptyCtx"), clientName, clientSecret).Return(client.Client{}, errors.New("failed to get client"))

	svc := client.NewService(cst.cfg, mockStore, &libcrypto.MockEd25519Generator{})

	_, err := svc.GetClient(context.Background(), clientName, clientSecret)
	cst.Require().Error(err)
}
