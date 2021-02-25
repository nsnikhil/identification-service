package resperr

import (
	"github.com/nsnikhil/erx"
	"net/http"
)

const (
	defaultStatusCode = http.StatusInternalServerError
	defaultMessage    = "internal server error"
)

func MapError(err error) ResponseError {
	t, ok := err.(*erx.Erx)
	if !ok {
		return NewResponseError(defaultStatusCode, defaultMessage)
	}

	k := t.Kind()

	switch k {
	case erx.ValidationError:
		return NewResponseError(http.StatusBadRequest, t.Error())
	case erx.ResourceNotFoundError:
		return NewResponseError(http.StatusNotFound, "resource not found")
	case erx.AuthenticationError:
		return NewResponseError(http.StatusUnauthorized, "authentication failed")
	case erx.InvalidCredentialsError:
		return NewResponseError(http.StatusUnauthorized, "invalid credentials")
	case erx.DuplicateRecordError:
		return NewResponseError(http.StatusConflict, "duplicate record")
	default:
		return NewResponseError(defaultStatusCode, defaultMessage)
	}

}
