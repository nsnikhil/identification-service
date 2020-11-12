package resperr_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/internal/resperr"
	"identification-service/pkg/liberr"
	"net/http"
	"testing"
)

func TestErrorMap(t *testing.T) {
	testCases := map[string]struct {
		err             error
		expectedRespErr resperr.ResponseError
	}{
		"test mapping for validation error": {
			err:             liberr.WithArgs(liberr.ValidationError, errors.New("invalid credentials")),
			expectedRespErr: resperr.NewResponseError(http.StatusBadRequest, "invalid credentials"),
		},
		"test mapping for resource not found error": {
			err:             liberr.WithArgs(liberr.ResourceNotFound, errors.New("not user found with id 1")),
			expectedRespErr: resperr.NewResponseError(http.StatusNotFound, "resource not found"),
		},
		"test mapping for authentication error": {
			err:             liberr.WithArgs(liberr.AuthenticationError, errors.New("authentication failed")),
			expectedRespErr: resperr.NewResponseError(http.StatusUnauthorized, "authentication failed"),
		},
		"test mapping for invalid credentials error": {
			err:             liberr.WithArgs(liberr.InvalidCredentialsError, errors.New("invalid user name or password")),
			expectedRespErr: resperr.NewResponseError(http.StatusUnauthorized, "invalid credentials"),
		},
		"test mapping for lib error with no kind": {
			err:             liberr.WithArgs(errors.New("database error")),
			expectedRespErr: resperr.NewResponseError(http.StatusInternalServerError, "internal server error"),
		},
		"test mapping for not lib error": {
			err:             errors.New("database error"),
			expectedRespErr: resperr.NewResponseError(http.StatusInternalServerError, "internal server error"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedRespErr, resperr.MapError(testCase.err))
		})
	}
}
