package liberr_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/liberr"
	"testing"
)

func TestNewErrorWithArgsSuccess(t *testing.T) {
	testCases := map[string][]interface{}{
		"test create error with error": {errors.New("some error")},
		"test create error with error, operations": {
			errors.New("some error"),
			liberr.Operation("some operation"),
		},
		"test create error with error, operations, severity": {
			errors.New("some error"),
			liberr.SeverityError,
			liberr.Operation("some operation"),
		},
		"test create error with error, operations, severity, kind": {
			errors.New("some error"),
			liberr.SeverityError,
			liberr.InternalError,
			liberr.Operation("some operation"),
		},
		"test create error with nested error": {
			liberr.WithArgs(errors.New("internal error")),
		},
	}

	for name, args := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.NotNil(t, liberr.WithArgs(args...))
		})
	}
}

func TestCreateNewErrorWithOp(t *testing.T) {
	testCases := map[string]struct {
		op  liberr.Operation
		err error
	}{
		"test create new error": {
			op:  liberr.Operation("some op"),
			err: errors.New("some error"),
		},
		"test create new nested error": {
			op:  liberr.Operation("some op"),
			err: liberr.WithArgs(errors.New("some error")),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.NotNil(t, liberr.WithOp(testCase.op, testCase.err))
		})
	}
}

func TestNewErrorWithArgsFailure(t *testing.T) {
	testCases := map[string][]interface{}{
		"test failure when args are empty":   {},
		"test failure when error is missing": {liberr.Operation("some op")},
		"test failure when a error if redundant": {
			errors.New("some error"),
			errors.New("other error"),
		},
		"test failure when a operation if redundant": {
			errors.New("some error"),
			liberr.Operation("some op"),
			liberr.Operation("some op"),
		},
		"test failure when a severity if redundant": {
			errors.New("some error"),
			liberr.SeverityError,
			liberr.SeverityError,
		},

		"test failure when a kind if redundant": {
			errors.New("some error"),
			liberr.InternalError,
			liberr.InternalError,
		},
	}

	for name, args := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Nil(t, liberr.WithArgs(args...))
		})
	}
}

func TestGetErrorKind(t *testing.T) {
	testCases := map[string]struct {
		err          *liberr.Error
		expectedKind liberr.Kind
	}{
		"test get kind of the top most error": {
			err:          liberr.WithArgs(errors.New("some error"), liberr.InternalError),
			expectedKind: liberr.InternalError,
		},
		"test get kind of the nested most error": {
			err: liberr.WithArgs(
				liberr.Operation("some op"),
				liberr.WithArgs(errors.New("some error"), liberr.InternalError),
			),
			expectedKind: liberr.InternalError,
		},
		"test return empty string when kind is missing": {
			err:          liberr.WithArgs(errors.New("some error")),
			expectedKind: "",
		},
		"test return empty string when err si nil": {
			err:          nil,
			expectedKind: "",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedKind, testCase.err.Kind())
		})
	}
}

func TestGetCause(t *testing.T) {
	testCases := map[string]struct {
		err   *liberr.Error
		cause error
	}{
		"test get cause for error": {
			err:   liberr.WithArgs(errors.New("some error")),
			cause: errors.New("some error"),
		},
		"test get cause for nested error": {
			err: liberr.WithArgs(
				liberr.Operation("some op"),
				liberr.WithArgs(errors.New("some error")),
			),
			cause: errors.New("some error"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.cause.Error(), testCase.err.Error())
		})
	}
}

func TestGetEncodedStack(t *testing.T) {
	testCases := map[string]struct {
		err            *liberr.Error
		expectedResult string
	}{
		"test get encoded stack with error": {
			err:            liberr.WithArgs(errors.New("some error")),
			expectedResult: "[cause:some error]",
		},
		"test get encoded stack with error and operation": {
			err: liberr.WithArgs(
				liberr.Operation("some op"),
				errors.New("some error"),
			),
			expectedResult: "[operation:some op  cause:some error]",
		},
		"test get encoded stack with error, operation and kind": {
			err: liberr.WithArgs(
				liberr.Operation("some op"),
				liberr.InternalError,
				errors.New("some error"),
			),
			expectedResult: "[kind:internalError  operation:some op  cause:some error]",
		},
		"test get encoded stack with error, operation, kind and severity": {
			err: liberr.WithArgs(
				liberr.Operation("some op"),
				liberr.InternalError,
				liberr.SeverityError,
				errors.New("some error"),
			),
			expectedResult: "[kind:internalError  operation:some op  severity:error  cause:some error]",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedResult, testCase.err.EncodedStack())
		})
	}
}
