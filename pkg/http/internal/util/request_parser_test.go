package util_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/http/internal/util"
	"identification-service/pkg/liberr"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestParseRequest(t *testing.T) {
	type CusReq struct {
		ReqID   string `json:"req_id"`
		ReqData string `json:"req_data"`
	}

	rf := func(method, path string, body io.Reader) *http.Request {
		req, err := http.NewRequest(method, path, body)
		require.NoError(t, err)

		return req
	}

	testCases := map[string]struct {
		name           string
		input          func() *http.Request
		expectedResult interface{}
		expectedError  error
	}{
		"test request parse success": {
			input: func() *http.Request {
				cr := CusReq{ReqID: "req-id", ReqData: "req data"}

				b, err := json.Marshal(&cr)
				require.NoError(t, err)

				return rf(http.MethodGet, "/random", bytes.NewBuffer(b))
			},
			expectedResult: CusReq{ReqID: "req-id", ReqData: "req data"},
		},
		"test request parse failure when req is nil": {
			input:          func() *http.Request { return nil },
			expectedResult: CusReq{},
			expectedError:  liberr.WithArgs(errors.New("request is nil")),
		},
		"test request parse failure when req body is nil": {
			input: func() *http.Request {
				return rf(http.MethodGet, "/random", nil)
			},
			expectedResult: CusReq{},
			expectedError:  liberr.WithArgs(errors.New("request body is nil")),
		},
		"test request parse failure when fail to read body": {
			input: func() *http.Request {
				cr := CusReq{ReqID: "req-id", ReqData: "req data"}

				b, err := json.Marshal(&cr)
				require.NoError(t, err)

				r := rf(http.MethodGet, "/random", bytes.NewBuffer(b))

				_, err = ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				return r
			},
			expectedResult: CusReq{},
			expectedError:  liberr.WithArgs(errors.New("unexpected end of JSON input")),
		},
		"test request parse failure when unmarshalling fails": {
			input: func() *http.Request {
				type CusReq struct {
					ReqID   string   `json:"req_id"`
					ReqData []string `json:"req_data"`
				}

				cr := CusReq{ReqID: "req-id", ReqData: []string{"req data"}}

				b, err := json.Marshal(&cr)
				require.NoError(t, err)

				return rf(http.MethodGet, "/random", bytes.NewBuffer(b))
			},
			expectedResult: CusReq{ReqID: "req-id"},
			expectedError:  errors.New("json: cannot unmarshal array into Go struct field CusReq.req_data of type string"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.input()

			var target CusReq

			err := util.ParseRequest(req, &target)

			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, testCase.expectedResult, target)
		})
	}
}
