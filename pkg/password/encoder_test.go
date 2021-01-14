package password_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/config"
	"identification-service/pkg/password"
	"identification-service/pkg/test"
	"testing"
)

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
	key := et.encoder.GenerateKey(test.NewPassword(), test.RandBytes(86))
	require.NotNil(et.T(), key)
}

func (et *encoderTestSuite) TestPasswordEncoderEncodeKey() {
	hash := et.encoder.EncodeKey(test.RandBytes(32))
	require.NotEmpty(et.T(), hash)
}

func (et *encoderTestSuite) TestPasswordEncoderValidatePasswordSuccess() {
	userPassword := test.NewPassword()

	salt, err := et.encoder.GenerateSalt()
	et.Require().NoError(err)

	key := et.encoder.GenerateKey(userPassword, salt)
	hash := et.encoder.EncodeKey(key)

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
			hash:         test.RandString(44),
			salt:         test.RandBytes(86),
		},
		"test failure when hash is different": {
			userPassword: test.NewPassword(),
			hash:         "OtherHash",
			salt:         test.RandBytes(86),
		},
		"test failure when salt is different": {
			userPassword: test.NewPassword(),
			hash:         test.RandString(44),
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
