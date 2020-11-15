package libcrypto_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"identification-service/pkg/libcrypto"
	"testing"
)

const encodedPem = `LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFBQUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUFNd0FBQUF0emMyZ3RaVwpReU5UVXhPUUFBQUNCZ2tpL3hrenYwdXpaL3JGUkRnVEt5dFh6RnpDVHluOTEyaVdRTHg3MlJsQUFBQUxDamNCc0FvM0FiCkFBQUFBQXR6YzJndFpXUXlOVFV4T1FBQUFDQmdraS94a3p2MHV6Wi9yRlJEZ1RLeXRYekZ6Q1R5bjkxMmlXUUx4NzJSbEEKQUFBRUJQUE0yNDhHL2VaZ1NpZUl1dUtQNG5YTVY4TmNSK1MybzhKM1Rsczl1SDVXQ1NML0dUTy9TN05uK3NWRU9CTXJLMQpmTVhNSlBLZjNYYUpaQXZIdlpHVUFBQUFKbTVwYTJocGJITnZibWxBVG1scmFHbHNjeTFOWVdOQ2IyOXJMVkJ5YnkweUxtCnh2WTJGc0FRSURCQVVHQnc9PQotLS0tLUVORCBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0K`

type keyGeneratorTestSuite struct {
	keyGenerator libcrypto.Ed25519Generator
	suite.Suite
}

func (kts *keyGeneratorTestSuite) SetupSuite() {
	kts.keyGenerator = libcrypto.NewKeyGenerator()
}

func (kts *keyGeneratorTestSuite) TestGenerateSuccess() {
	_, _, err := kts.keyGenerator.Generate()
	fmt.Println(kts.keyGenerator.Generate())
	assert.Nil(kts.T(), err)
}

func (kts *keyGeneratorTestSuite) TestFromPemSuccess() {
	_, _, err := kts.keyGenerator.FromEncodedPem(encodedPem)
	assert.Nil(kts.T(), err)
}

func (kts *keyGeneratorTestSuite) TestFromPemFailure() {
	_, _, err := kts.keyGenerator.FromEncodedPem("invalidPem")
	assert.Error(kts.T(), err)
}

func TestKeyGenerator(t *testing.T) {
	suite.Run(t, new(keyGeneratorTestSuite))
}
