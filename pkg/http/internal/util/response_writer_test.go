package util_test

import (
	"github.com/stretchr/testify/assert"
	"identification-service/pkg/http/internal/resperr"
	"identification-service/pkg/http/internal/util"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteSuccessResponse(t *testing.T) {
	testCases := []struct {
		name           string
		actualResult   func() (string, int)
		expectedCode   int
		expectedResult string
	}{
		{
			name: "write success response success",
			actualResult: func() (string, int) {
				type CusResp struct {
					RespID   string `json:"resp_id"`
					RespData string `json:"resp_data"`
				}

				cr := CusResp{"resp-id", "resp data"}

				w := httptest.NewRecorder()

				util.WriteSuccessResponse(http.StatusCreated, cr, w)

				return w.Body.String(), w.Code
			},
			expectedCode:   http.StatusCreated,
			expectedResult: "{\"data\":{\"resp_id\":\"resp-id\",\"resp_data\":\"resp data\"},\"success\":true}",
		},
		{
			name: "write success response failure",
			actualResult: func() (string, int) {
				w := httptest.NewRecorder()

				c := make(chan int)

				util.WriteSuccessResponse(http.StatusCreated, c, w)

				return w.Body.String(), w.Code
			},
			expectedCode:   http.StatusInternalServerError,
			expectedResult: "internal server error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res, code := testCase.actualResult()

			assert.Equal(t, testCase.expectedCode, code)
			assert.Equal(t, testCase.expectedResult, res)
		})
	}
}

func TestWriteFailureResponse(t *testing.T) {
	testCases := []struct {
		name           string
		actualResult   func() (string, int)
		expectedCode   int
		expectedResult string
	}{
		{
			name: "write failure response success",
			actualResult: func() (string, int) {
				err := resperr.NewResponseError(http.StatusBadRequest, "failed to parse")

				w := httptest.NewRecorder()

				util.WriteFailureResponse(err, w)

				return w.Body.String(), w.Code
			},
			expectedCode:   http.StatusBadRequest,
			expectedResult: "{\"error\":{\"message\":\"failed to parse\"},\"success\":false}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res, code := testCase.actualResult()

			assert.Equal(t, testCase.expectedCode, code)
			assert.Equal(t, testCase.expectedResult, res)
		})
	}
}
