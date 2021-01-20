package router_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/http/router"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"identification-service/pkg/user"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	mockConfig := &config.MockConfig{}
	mockConfig.On("Env").Return("dev")
	mockConfig.On("AuthConfig").Return(config.AuthConfig{})

	r := router.NewRouter(
		mockConfig, &reporters.MockLogger{}, &reporters.MockPrometheus{},
		&client.MockService{}, &user.MockService{}, &session.MockService{},
	)

	rf := func(method, path string) *http.Request {
		req, err := http.NewRequest(method, path, nil)
		require.NoError(t, err)
		return req
	}

	testCases := map[string]struct {
		name    string
		request *http.Request
	}{
		"test ping route": {
			request: rf(http.MethodGet, "/ping"),
		},
		"test metrics route": {
			request: rf(http.MethodGet, "/metrics"),
		},
		"test create user route": {
			request: rf(http.MethodPost, "/user/sign-up"),
		},
		"test update password route": {
			request: rf(http.MethodPost, "/user/update-password"),
		},
		"test session login route": {
			request: rf(http.MethodPost, "/session/login"),
		},
		"test session refresh token route": {
			request: rf(http.MethodPost, "/session/refresh-token"),
		},
		"test session logout route": {
			request: rf(http.MethodPost, "/session/logout"),
		},
		"test client register route": {
			request: rf(http.MethodPost, "/client/register"),
		},
		"test client revoke route": {
			request: rf(http.MethodPost, "/client/revoke"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r.ServeHTTP(w, testCase.request)

			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}
