package util_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/util"
	"testing"
)

func TestIsValidUUIDSuccess(t *testing.T) {
	testIsValidUUID(t, true, "86d690dd-92a0-40ac-ad48-110c951e3cb8")
}

func TestIsValidUUIDFailure(t *testing.T) {
	testIsValidUUID(t, false, "invalidUserID")
}

func testIsValidUUID(t *testing.T, isValid bool, id string) {
	assert.Equal(t, isValid, util.IsValidUUID(id))
}
