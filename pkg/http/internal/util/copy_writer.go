package util

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type CopyWriter struct {
	http.ResponseWriter
	data *bytes.Buffer
	code int
}

func NewCopyWriter(resp http.ResponseWriter) *CopyWriter {
	return &CopyWriter{
		ResponseWriter: resp,
		data:           new(bytes.Buffer),
	}
}

func (cw *CopyWriter) Write(b []byte) (int, error) {
	n, err := cw.data.Write(b)
	if err != nil {
		return n, err
	}

	return cw.ResponseWriter.Write(b)
}

func (cw *CopyWriter) WriteHeader(statusCode int) {
	cw.code = statusCode
	cw.ResponseWriter.WriteHeader(statusCode)
}

func (cw *CopyWriter) Body() ([]byte, error) {
	b, err := ioutil.ReadAll(cw.data)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (cw *CopyWriter) Code() int {
	return cw.code
}
