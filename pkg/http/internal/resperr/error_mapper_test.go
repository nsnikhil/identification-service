package resperr_test

import (
	"errors"
	"github.com/nsnikhil/erx"
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/internal/resperr"
	"net/http"
	"testing"
)

func TestErrorMap(t *testing.T) {
	testCases := map[string]struct {
		err             error
		expectedRespErr resperr.ResponseError
	}{
		"test mapping for validation error": {
			err:             erx.WithArgs(erx.ValidationError, errors.New("invalid credentials")),
			expectedRespErr: resperr.NewResponseError(http.StatusBadRequest, "invalid credentials"),
		},
		"test mapping for resource not found error": {
			err:             erx.WithArgs(erx.ResourceNotFoundError, errors.New("not user found with id 1")),
			expectedRespErr: resperr.NewResponseError(http.StatusNotFound, "resource not found"),
		},
		"test mapping for authentication error": {
			err:             erx.WithArgs(erx.AuthenticationError, errors.New("authentication failed")),
			expectedRespErr: resperr.NewResponseError(http.StatusUnauthorized, "authentication failed"),
		},
		"test mapping for invalid credentials error": {
			err:             erx.WithArgs(erx.InvalidCredentialsError, errors.New("invalid user name or password")),
			expectedRespErr: resperr.NewResponseError(http.StatusUnauthorized, "invalid credentials"),
		},
		"test mapping for duplicate record error": {
			err:             erx.WithArgs(erx.DuplicateRecordError, errors.New("duplicate record")),
			expectedRespErr: resperr.NewResponseError(http.StatusConflict, "duplicate record"),
		},
		"test mapping for lib error with no kind": {
			err:             erx.WithArgs(errors.New("database error")),
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
