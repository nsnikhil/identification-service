package test

import (
	"crypto/ed25519"
	cr "crypto/rand"
	"fmt"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const (
	EmptyString = ""
	Zero        = 0
	QueueName   = "test_queue"
	letters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	QueryTTL    = 10000
)

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())

	sz := len(letters)

	sb := strings.Builder{}
	sb.Grow(n)

	for i := 0; i < n; i++ {
		sb.WriteByte(letters[rand.Intn(sz)])
	}

	return sb.String()
}

func randBytes(n int) []byte {
	res := make([]byte, n)
	rand.Read(res)
	return res
}

var ClientTableName = "clients"
var ClientName = func() string { return fmt.Sprintf("client%s", randString(8)) }
var ClientAccessTokenTTL = 10
var ClientSessionTTL = 87601
var ClientMaxActiveSessions = 2
var ClientSessionStrategyRevokeOld = "revoke_old"
var ClientID = func() string { return uuid.New().String() }
var ClientSecret = func() string { return uuid.New().String() }
var ClientEncodedPublicKey = func() string { return randString(44) }

var UserTableName = "users"
var UserName = func() string { return fmt.Sprintf("Test %s", randString(8)) }
var UserEmail = func() string { return fmt.Sprintf("%s@mail.com", randString(8)) }
var UserPassword = "Password@1234"
var UserPasswordNew = "NewPassword@1234"
var UserPasswordInvalid = "password@1234"
var UserID = func() string { return uuid.New().String() }
var UserPasswordHash = func() string { return randString(44) }
var UserPasswordSalt = func() []byte { return randBytes(86) }
var UserPasswordKey = func() []byte { return randBytes(32) }

var SessionTableName = "sessions"
var SessionID = func() string { return uuid.New().String() }
var SessionAccessToken = func() string { return generateToken() }
var SessionRefreshToken = func() string { return uuid.New().String() }

var NewClient = func(t *testing.T) client.Client {
	cl, err := client.NewClientBuilder().
		Name(ClientName()).
		AccessTokenTTL(ClientAccessTokenTTL).
		SessionTTL(ClientSessionTTL).
		SessionStrategy(ClientSessionStrategyRevokeOld).
		MaxActiveSessions(ClientMaxActiveSessions).
		PrivateKey(ClientPriKey()).
		Build()

	require.NoError(t, err)

	return cl
}

var CreatedAt = time.Date(2020, 11, 23, 23, 45, 0, 0, time.UTC)
var UpdatedAt = time.Date(2020, 11, 23, 23, 45, 0, 0, time.UTC)

var generateToken = func() string {
	key, err := paseto.NewV2().Sign(ClientPriKey(), paseto.JSONToken{}, nil)
	if err != nil {
		log.Fatal(err)
	}

	return key
}

var generateKey = func() (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, pri, err := ed25519.GenerateKey(cr.Reader)
	if err != nil {
		log.Fatal(err)
	}

	return pub, pri
}

var GenerateKey = func() (ed25519.PublicKey, ed25519.PrivateKey) {
	return generateKey()
}

var ClientPubKey = func() ed25519.PublicKey {
	key, _ := generateKey()
	return key
}

var ClientPriKey = func() ed25519.PrivateKey {
	_, key := generateKey()
	return key
}

var ClientPriKeyBytes = func() []byte {
	return ClientPriKey()
}
