package test

import (
	"crypto/ed25519"
	cr "crypto/rand"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"log"
	"math/rand"
	"strings"
	"time"
)

const (
	EmptyString                    = ""
	Zero                           = 0
	letters                        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers                        = "0123456789"
	symbols                        = "!@#$%^&*()"
	QueryTTL                       = 10000
	ClientTableName                = "clients"
	ClientSessionStrategyRevokeOld = "revoke_old"
	UserTableName                  = "users"
	SessionTableName               = "sessions"

	ClientIdKey                  = "id"
	ClientNameKey                = "name"
	ClientSecretKey              = "secret"
	ClientRevokedKey             = "revoked"
	ClientAccessTokenTTLKey      = "accessTokenTTL"
	ClientSessionTTLKey          = "sessionTTL"
	ClientMaxActiveSessionsKey   = "maxActiveSessions"
	ClientSessionStrategyNameKey = "sessionStrategyName"
	ClientPrivateKeyKey          = "privateKey"
	ClientCreatedAtKey           = "createdAt"
	ClientUpdatedAtKey           = "updatedAt"
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

var NewClient = func(cfg config.ClientConfig, d map[string]interface{}) (client.Client, error) {
	either := func(a interface{}, b interface{}) interface{} {
		if a == nil {
			return b
		}

		return a
	}

	return client.NewClientBuilder(cfg).
		ID(either(d[ClientIdKey], NewUUID()).(string)).
		Name(either(d[ClientNameKey], RandString(8)).(string)).
		Secret(either(d[ClientSecretKey], NewUUID()).(string)).
		Revoked(either(d[ClientRevokedKey], false).(bool)).
		AccessTokenTTL(either(d[ClientAccessTokenTTLKey], RandInt(1, 10)).(int)).
		SessionTTL(either(d[ClientSessionTTLKey], RandInt(1440, 86701)).(int)).
		MaxActiveSessions(either(d[ClientMaxActiveSessionsKey], RandInt(1, 10)).(int)).
		SessionStrategy(either(d[ClientSessionStrategyNameKey], ClientSessionStrategyRevokeOld).(string)).
		PrivateKey(either(d[ClientPrivateKeyKey], ClientPriKeyBytes()).([]byte)).
		CreatedAt(either(d[ClientCreatedAtKey], CreatedAt).(time.Time)).
		UpdatedAt(either(d[ClientUpdatedAtKey], UpdatedAt).(time.Time)).
		Build()
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
