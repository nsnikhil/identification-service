package test

import (
	"crypto/ed25519"
	cr "crypto/rand"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const (
	EmptyString                    = ""
	Zero                           = 0
	QueueName                      = "test_queue"
	letters                        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers                        = "0123456789"
	symbols                        = "!@#$%^&*()"
	QueryTTL                       = 10000
	ClientTableName                = "clients"
	ClientSessionStrategyRevokeOld = "revoke_old"
	UserTableName                  = "users"
	SessionTableName               = "sessions"
)

func RandString(n int) string {
	return randStringFrom(n, letters)
}

func randStringFrom(n int, values string) string {
	rand.Seed(time.Now().UnixNano())

	sz := len(values)

	sb := strings.Builder{}
	sb.Grow(n)

	for i := 0; i < n; i++ {
		sb.WriteByte(values[rand.Intn(sz)])
	}

	return sb.String()
}

func NewUUID() string {
	return uuid.New().String()
}

func RandInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min+1) + min
}

func RandBytes(n int) []byte {
	res := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	rand.Read(res)
	return res
}

func NewPassword() string {
	sb := strings.Builder{}

	sb.WriteString(randStringFrom(4, letters[26:]))
	sb.WriteString(randStringFrom(1, symbols))
	sb.WriteString(randStringFrom(1, numbers))
	sb.WriteString(randStringFrom(4, letters[:26]))

	return sb.String()
}

func NewEmail() string {
	sb := strings.Builder{}

	sb.WriteString(randStringFrom(6, letters))
	sb.WriteString("@mail.com")

	return sb.String()
}

type ClientData struct {
	Name             string
	AccessTokenTTL   int
	SessionTokenTTL  int
	SessionStrategy  int
	MaxActiveSession int
	PrivateKey       []byte
}

var NewClient = func(t *testing.T, data ...ClientData) client.Client {
	either := func(a, b interface{}) interface{} {
		switch v := a.(type) {
		case string:
			if len(v) != 0 {
				return a
			}
		case int:
			if v != 0 {
				return a
			}
		case []byte:
			if v != nil && len(v) != 0 {
				return a
			}
		}

		return b
	}

	if len(data) == 0 {
		data = make([]ClientData, 1)
	}

	cl, err := client.NewClientBuilder(config.NewConfig("../../local.env").ClientConfig()).
		Name(either(data[0].Name, RandString(8)).(string)).
		AccessTokenTTL(either(data[0].AccessTokenTTL, RandInt(1, 10)).(int)).
		SessionTTL(either(data[0].SessionTokenTTL, RandInt(1440, 86701)).(int)).
		SessionStrategy(ClientSessionStrategyRevokeOld).
		MaxActiveSessions(either(data[0].MaxActiveSession, RandInt(1, 10)).(int)).
		PrivateKey(either(data[0].PrivateKey, ClientPriKey()).(ed25519.PrivateKey)).
		Build()

	require.NoError(t, err)

	return cl
}

var CreatedAt = time.Date(2020, 11, 23, 23, 45, 0, 0, time.UTC)
var UpdatedAt = time.Date(2020, 11, 23, 23, 45, 0, 0, time.UTC)

var NewPasetoToken = func() string {
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
