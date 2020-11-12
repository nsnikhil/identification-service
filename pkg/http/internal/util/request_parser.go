package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"identification-service/pkg/liberr"
	"io/ioutil"
	"net/http"
)

func ParseRequest(req *http.Request, data interface{}) error {
	if req == nil {
		return e("", errors.New("request is nil"))
	}

	if req.Body == nil {
		return e("", errors.New("request body is nil"))
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return e("ioutil.ReadAll", err)
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return e("json.Unmarshal", err)
	}

	return nil
}

//TODO: REMOVE THIS HELPER FUNCTION OR AT-LEAST RENAME
func e(op string, err error) *liberr.Error {
	opf := func() liberr.Operation {
		if len(op) == 0 {
			return "ParseRequest"
		}
		return liberr.Operation(fmt.Sprintf("ParseRequest.%s", op))
	}

	return liberr.WithArgs(liberr.SeverityError, liberr.ValidationError, opf(), err)
}
