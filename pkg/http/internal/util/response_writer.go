package util

import (
	"encoding/json"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/resperr"
	"net/http"
)

func writeResponse(code int, data []byte, resp http.ResponseWriter) {
	resp.WriteHeader(code)
	//TODO: DEAL WITH ERROR FROM WRITE
	_, _ = resp.Write(data)
}

func writeAPIResponse(code int, ar contract.APIResponse, resp http.ResponseWriter) {
	b, err := json.Marshal(&ar)
	if err != nil {
		//TODO: SHOULD YOU WRITE INTERNAL SERVER ERROR WHEN MARSHALLING FAILS
		writeResponse(http.StatusInternalServerError, []byte("internal server error"), resp)
		return
	}

	writeResponse(code, b, resp)
}

func WriteSuccessResponse(statusCode int, data interface{}, resp http.ResponseWriter) {
	writeAPIResponse(statusCode, contract.NewSuccessResponse(data), resp)
}

func WriteFailureResponse(gr resperr.ResponseError, resp http.ResponseWriter) {
	writeAPIResponse(gr.StatusCode(), contract.NewFailureResponse(gr.Description()), resp)
}
