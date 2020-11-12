package password_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/password"
	"testing"
)

const userPassword = "Password@1234"

var salt = []byte{207, 197, 172, 211, 25, 191, 32, 226, 29, 208, 91, 100, 53, 122, 163, 139, 0, 72, 228, 249, 203, 229, 79, 57, 210, 227, 163, 212, 181, 84, 166, 28, 41, 139, 224, 49, 157, 70, 159, 6, 106, 67, 231, 114, 63, 138, 72, 127, 5, 127, 33, 89, 44, 253, 94, 43, 82, 239, 9, 9, 87, 82, 182, 63, 176, 189, 128, 146, 67, 63, 150, 99, 233, 184, 99, 34, 49, 113, 102, 127, 206, 0, 86, 28, 57, 205}
var key = []byte{179, 18, 161, 153, 20, 84, 124, 47, 146, 60, 11, 23, 22, 100, 62, 174, 188, 197, 217, 79, 210, 72, 87, 249, 94, 251, 243, 121, 189, 160, 139, 241}
var hash = "sxKhmRRUfC+SPAsXFmQ+rrzF2U/SSFf5Xvvzeb2gi/E="

type encoderTestSuite struct {
	encoder password.Encoder
	suite.Suite
}

func (et *encoderTestSuite) SetupSuite() {
	var cfg = config.NewConfig("../../local.env").PasswordConfig()
	et.encoder = password.NewEncoder(cfg)
}

func (et *encoderTestSuite) TestPasswordEncoderGenerateSalt() {
	salt, err := et.encoder.GenerateSalt()

	require.NoError(et.T(), err)
	require.NotNil(et.T(), salt)
}

func (et *encoderTestSuite) TestPasswordEncoderGenerateKey() {
	key := et.encoder.GenerateKey(userPassword, salt)
	require.NotNil(et.T(), key)
}

func (et *encoderTestSuite) TestPasswordEncoderEncodeKey() {
	hash := et.encoder.EncodeKey(key)
	require.NotEmpty(et.T(), hash)
}

func (et *encoderTestSuite) TestPasswordEncoderValidatePasswordSuccess() {
	require.NoError(et.T(), et.encoder.VerifyPassword(userPassword, hash, salt))
}

func (et *encoderTestSuite) TestPasswordEncoderValidatePasswordFailure() {
	testCases := map[string]struct {
		userPassword string
		hash         string
		salt         []byte
	}{
		"test failure when password is different": {
			userPassword: "OtherPassword@1234",
			hash:         hash,
			salt:         salt,
		},
		"test failure when hash is different": {
			userPassword: userPassword,
			hash:         "OtherHash",
			salt:         salt,
		},
		"test failure when salt is different": {
			userPassword: userPassword,
			hash:         hash,
			salt:         []byte{1, 2, 3, 4},
		},
	}

	for name, testCase := range testCases {
		et.T().Run(name, func(t *testing.T) {
			require.Error(t, et.encoder.VerifyPassword(testCase.userPassword, testCase.hash, testCase.salt))
		})
	}
}

func (et *encoderTestSuite) TestValidatePasswordSuccess() {
	assert.Nil(et.T(), et.encoder.ValidatePassword("Password@1234"))
}

func (et *encoderTestSuite) TestValidatePasswordFailure() {
	testCases := map[string]string{
		"invalidPasswordOne  ": "password@1234",
		"invalidPasswordTwo  ": "PASSWORD@1234",
		"invalidPasswordThree": "Password@",
		"invalidPasswordFour ": "Password1",
		"invalidPasswordFive ": "Pa@1",
	}

	for name, givenPassword := range testCases {
		et.Run(name, func() {
			assert.Error(et.T(), et.encoder.ValidatePassword(givenPassword))
		})
	}
}

func TestEncoder(t *testing.T) {
	suite.Run(t, new(encoderTestSuite))
}
