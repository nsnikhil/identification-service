package router

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/http/internal/handler"
	mdl "identification-service/pkg/http/internal/middleware"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"identification-service/pkg/user"
	"net/http"
)

func NewRouter(cfg config.Config, lgr *zap.Logger, pr reporters.Prometheus, cs client.Service, us user.Service, ss session.Service) http.Handler {
	return getChiRouter(cfg, lgr, pr, cs, us, ss)
}

//TODO: FIX MIDDLEWARE REPETITION CODE
func getChiRouter(cfg config.Config, lgr *zap.Logger, pr reporters.Prometheus, cs client.Service, us user.Service, ss session.Service) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/ping", mdl.WithResponseHeaders(handler.PingHandler()))
	r.Handle("/metrics", promhttp.Handler())

	registerUserRoutes(r, lgr, pr, cs, us)
	registerSessionRoutes(r, lgr, pr, cs, ss)
	registerClientRoutes(r, cfg.AuthConfig(), lgr, pr, cs)

	return r
}

func registerUserRoutes(r chi.Router, lgr *zap.Logger, pr reporters.Prometheus, cs client.Service, us user.Service) {
	uh := handler.NewUserHandler(us)

	signUpHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("user", "sign-up"),
				mdl.WithClientAuth(lgr, cs,
					mdl.WithErrorHandler(lgr, uh.SignUp)),
			),
		),
	)

	updatePasswordHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("user", "update-password"),
				mdl.WithClientAuth(lgr, cs,
					mdl.WithErrorHandler(lgr, uh.UpdatePassword)),
			),
		),
	)

	r.Route("/user", func(r chi.Router) {
		r.Post("/sign-up", signUpHandler)
		r.Post("/update-password", updatePasswordHandler)
	})
}

func registerSessionRoutes(r chi.Router, lgr *zap.Logger, pr reporters.Prometheus, cs client.Service, ss session.Service) {
	sh := handler.NewSessionHandler(ss)

	loginHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("session", "login"),
				mdl.WithClientAuth(lgr, cs,
					mdl.WithErrorHandler(lgr, sh.Login)),
			),
		),
	)

	refreshTokenHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("session", "refresh-token"),
				mdl.WithClientAuth(lgr, cs,
					mdl.WithErrorHandler(lgr, sh.RefreshToken)),
			),
		),
	)

	logoutHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("session", "logout"),
				mdl.WithClientAuth(lgr, cs,
					mdl.WithErrorHandler(lgr, sh.Logout)),
			),
		),
	)

	r.Route("/session", func(r chi.Router) {
		r.Post("/login", loginHandler)
		r.Post("/refresh-token", refreshTokenHandler)
		r.Post("/logout", logoutHandler)
	})
}

func registerClientRoutes(r chi.Router, cfg config.AuthConfig, lgr *zap.Logger, pr reporters.Prometheus, ss client.Service) {
	ch := handler.NewClientHandler(ss)

	cred := map[string]string{cfg.UserName(): cfg.Password()}

	registerHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("client", "register"),
				mdl.WithBasicAuth(cred, lgr, "client",
					mdl.WithErrorHandler(lgr, ch.Register)),
			),
		),
	)

	revokeHandler := mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(pr, apiFunc("client", "revoke"),
				mdl.WithBasicAuth(cred, lgr, "client",
					mdl.WithErrorHandler(lgr, ch.Revoke)),
			),
		),
	)

	r.Route("/client", func(r chi.Router) {
		r.Post("/register", registerHandler)
		r.Post("/revoke", revokeHandler)
	})
}

func apiFunc(api, path string) string {
	return fmt.Sprintf("%s_%s", api, path)
}
