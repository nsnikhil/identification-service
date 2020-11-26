package middleware

import (
	"context"
	"errors"
	"fmt"
	"identification-service/pkg/client"
	"identification-service/pkg/http/internal/resperr"
	"identification-service/pkg/http/internal/util"
	"identification-service/pkg/liberr"
	reporters "identification-service/pkg/reporting"
	"net/http"
	"time"
)

func WithErrorHandler(lgr reporters.Logger, handler func(resp http.ResponseWriter, req *http.Request) error) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		err := handler(resp, req)
		if err == nil {
			return
		}

		logAndWriteError(lgr, resp, err)
	}
}

//TODO: ADD MASKING BEFORE LOGGING REQ AND RESP
func WithReqRespLog(lgr reporters.Logger, handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		cr := util.NewCopyWriter(resp)

		handler(cr, req)

		//b, _ := cr.Body()

		//lgr.Sugar().Debug(req)
		//lgr.Sugar().Debug(string(b))
	}
}

func WithResponseHeaders(handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		headers := map[string]string{
			"Content-Type":              "application/json",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
			"X-Frame-Options":           "deny",
			"X-Content-Type-Options":    "nosniff",
		}

		for key, value := range headers {
			resp.Header().Set(key, value)
		}

		handler(resp, req)
	}
}

func WithBasicAuth(cred map[string]string, lgr reporters.Logger, realm string, handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {

		authFailed := func(resp http.ResponseWriter, realm string) {
			resp.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))

			logAndWriteError(lgr, resp, liberr.WithArgs(
				liberr.Operation("WithBasicAuth"),
				liberr.AuthenticationError,
				errors.New("basic auth failed"),
			))
		}

		userName, password, ok := req.BasicAuth()
		if !ok {
			authFailed(resp, realm)
			return
		}

		credPass, credUserOk := cred[userName]
		if !credUserOk || password != credPass {
			authFailed(resp, realm)
			return
		}

		handler(resp, req)
	}
}

func WithClientAuth(lgr reporters.Logger, service client.Service, handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		name := req.Header.Get("CLIENT-ID")
		secret := req.Header.Get("CLIENT-SECRET")

		cl, err := service.GetClient(req.Context(), name, secret)
		if err != nil {
			logAndWriteError(lgr, resp, liberr.WithArgs(
				liberr.Operation("WithClientAuth"),
				liberr.AuthenticationError,
				err,
			))
			return
		}

		if cl.IsRevoked() {
			logAndWriteError(lgr, resp, liberr.WithArgs(
				liberr.Operation("WithClientAuth"),
				liberr.AuthenticationError,
				errors.New("client revoked"),
			))
			return
		}

		ctx, err := client.WithContext(req.Context(), cl)
		if err != nil {
			logAndWriteError(lgr, resp, liberr.WithOp("WithClientAuth", err))
			return
		}

		handler(resp, req.WithContext(ctx))
	}
}

func logAndWriteError(lgr reporters.Logger, resp http.ResponseWriter, err error) {
	t, ok := err.(*liberr.Error)
	if ok {
		lgr.Error(t.EncodedStack())
	} else {
		lgr.Error(err.Error())
	}

	util.WriteFailureResponse(resperr.MapError(err), resp)
}

func WithRequestContext(handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		// TODO: CHANGE TEMP VALUE
		ctx := context.WithValue(req.Context(), "key", "val")
		handler(resp, req.WithContext(ctx))
	}
}

func WithPrometheus(prometheus reporters.Prometheus, api string, handler http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		// TODO CHANGE THIS
		hasError := func(code int) bool {
			return code >= 400 && code <= 600
		}

		start := time.Now()
		prometheus.ReportAttempt(api)

		cr := util.NewCopyWriter(resp)

		handler(cr, req)
		if hasError(cr.Code()) {
			duration := time.Since(start)
			prometheus.Observe(api, duration.Seconds())
			prometheus.ReportFailure(api)
			return
		}

		duration := time.Since(start)
		prometheus.Observe(api, duration.Seconds())

		prometheus.ReportSuccess(api)
	}
}
