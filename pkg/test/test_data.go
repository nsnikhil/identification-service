package test

import (
	"crypto/ed25519"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"testing"
	"time"
)

const (
	EmptyString = ""
	Zero        = 0
	QueueName   = "test_queue"

	//CLIENT
	ClientTableName                = "clients"
	ClientName                     = "clientOne"
	ClientAccessTokenTTL           = 10
	ClientSessionTTL               = 87601
	ClientMaxActiveSessions        = 2
	ClientSessionStrategyRevokeOld = "revoke_old"
	ClientID                       = "f14abb31-ec1a-4ff6-a937-c2e930ca34ef"
	ClientSecret                   = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
	ClientEncodedPublicKey         = "8lchzCKRbdXEHsG/hJNMjMqdJLbIvAvDoViJtlcwWWo"

	//USER
	UserTableName       = "users"
	UserName            = "Test Name"
	UserEmail           = "test@test.com"
	UserPassword        = "Password@1234"
	UserPasswordNew     = "NewPassword@1234"
	UserPasswordInvalid = "password@1234"
	UserID              = "86d690dd-92a0-40ac-ad48-110c951e3cb8"
	UserPasswordHash    = "IrNrmgBeMAGGLIB//hF8+EVgxK6S/4NbXo9pIeadTfM="

	//SESSION
	SessionTableName         = "sessions"
	SessionID                = "f113fe5c-de2f-4876-b734-b51fbdc96e4b"
	SessionAccessToken       = "v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMDozNjowNyswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTA6MjY6MDcrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiMTEwMTI0NjUtMDNhNC00OWI2LTgwODEtY2RmYzczMDlhY2MwIiwibmJmIjoiMjAyMC0xMS0wN1QxMDoyNjowNyswNTozMCJ9PrXViH5779NxXHK_PxnwW-FdFV0klU07umd8X7F0A9irFLX7GTS3AczNm_hmb_yfYOX0o4DJri89AWeCb0qTAg.bnVsbA"
	SessionAccessTokenTwo    = "v2.public.eyJhdWQiOiJ1c2VyIiwiZXhwIjoiMjAyMC0xMS0wN1QxMjozNDowOCswNTozMCIsImlhdCI6IjIwMjAtMTEtMDdUMTI6MjQ6MDgrMDU6MzAiLCJpc3MiOiJpZGVudGlmaWNhdGlvbi1zZXJ2aWNlIiwianRpIjoiZjJiNzhlNWYtNTZhMi00MzMwLWFhYWUtYmM4OWM1NzllNzIwIiwibmJmIjoiMjAyMC0xMS0wN1QxMjoyNDowOCswNTozMCIsInN1YiI6Ijg2ZDY5MGRkLTkyYTAtNDBhYy1hZDQ4LTExMGM5NTFlM2NiOCJ9DHCzvrlz6_QDB6zuuQcAmZs6yFoqBgkcHbtIVRcsDJ068XGs6N5R4U069lQvy-r7fHY2pL6tmxjRAZq1McetAA.bnVsbA"
	SessionRefreshToken      = "5df8159e-fd51-4e6c-9849-a9b1f070a403"
	SessionRefreshTokenTwo   = "7c70c869-df19-45f3-a51c-23d88d2ff5d5"
	SessionRefreshTokenThree = "0798a3b4-e090-4dad-8f92-024260028bd5"
	SessionRefreshTokenFour  = "3e99c610-bf55-45ff-8d8d-0f11b0c9b4ce"
)

var NewClient = func(t *testing.T) client.Client {
	cl, err := client.NewClientBuilder().
		Name(ClientName).
		AccessTokenTTL(ClientAccessTokenTTL).
		SessionTTL(ClientSessionTTL).
		SessionStrategy(ClientSessionStrategyRevokeOld).
		MaxActiveSessions(ClientMaxActiveSessions).
		PrivateKey(ClientPriKey).
		Build()

	require.NoError(t, err)

	return cl
}

var CreatedAt = time.Date(2020, 11, 23, 23, 45, 0, 0, time.UTC)
var UpdatedAt = time.Date(2020, 11, 23, 23, 45, 0, 0, time.UTC)

var QueryTTL = 1000

var UserPasswordSalt = []byte{90, 20, 247, 194, 220, 48, 153, 58, 158, 103, 9, 17, 243, 24, 179, 254, 88, 59, 161, 81, 216, 8, 126, 122, 102, 151, 200, 12, 134, 118, 146, 197, 193, 248, 117, 57, 127, 137, 112, 233, 116, 50, 128, 84, 127, 93, 180, 23, 81, 69, 245, 183, 45, 57, 51, 125, 9, 46, 200, 175, 97, 49, 11, 0, 40, 228, 186, 60, 177, 43, 69, 52, 168, 195, 69, 101, 21, 245, 62, 131, 252, 96, 240, 154, 251, 2}
var UserPasswordKey = []byte{34, 179, 107, 154, 0, 94, 48, 1, 134, 44, 128, 127, 254, 17, 124, 248, 69, 96, 196, 174, 146, 255, 131, 91, 94, 143, 105, 33, 230, 157, 77, 243}

var ClientPubKey = ed25519.PublicKey{6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}
var ClientPriKey = ed25519.PrivateKey{3, 195, 208, 247, 190, 104, 63, 62, 164, 50, 63, 217, 229, 215, 179, 62, 223, 104, 197, 43, 164, 164, 231, 1, 22, 70, 154, 130, 109, 98, 88, 210, 6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}
var ClientPriKeyBytes = []byte{3, 195, 208, 247, 190, 104, 63, 62, 164, 50, 63, 217, 229, 215, 179, 62, 223, 104, 197, 43, 164, 164, 231, 1, 22, 70, 154, 130, 109, 98, 88, 210, 6, 170, 181, 226, 181, 81, 223, 119, 104, 220, 249, 120, 90, 158, 6, 10, 117, 97, 114, 150, 55, 185, 206, 184, 47, 231, 164, 120, 137, 75, 184, 216}
