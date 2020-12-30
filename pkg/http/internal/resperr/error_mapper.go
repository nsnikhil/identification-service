package resperr

import (
	"identification-service/pkg/liberr"
	"net/http"
)

const (
	defaultStatusCode = http.StatusInternalServerError
	defaultMessage    = "internal server error"
)

func MapError(err error) ResponseError {
	t, ok := err.(*liberr.Error)
	if !ok {
		return NewResponseError(defaultStatusCode, defaultMessage)
	}

	k := t.Kind()

	switch k {
	case liberr.ValidationError:
		return NewResponseError(http.StatusBadRequest, t.Error())
	case liberr.ResourceNotFound:
		return NewResponseError(http.StatusNotFound, "resource not found")
	case liberr.AuthenticationError:
		return NewResponseError(http.StatusUnauthorized, "authentication failed")
	case liberr.InvalidCredentialsError:
		return NewResponseError(http.StatusUnauthorized, "invalid credentials")
	case liberr.DuplicateRecordError:
		return NewResponseError(http.StatusConflict, "duplicate record")
	default:
		return NewResponseError(defaultStatusCode, defaultMessage)
	}

}
