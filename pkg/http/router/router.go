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

func NewRouter(cfg config.Config, lgr *zap.Logger, prometheus reporters.Prometheus, clientService client.Service, userService user.Service, sessionService session.Service) http.Handler {
	return getChiRouter(cfg, lgr, prometheus, userService, sessionService, clientService)
}

func getChiRouter(cfg config.Config, lgr *zap.Logger, pr reporters.Prometheus, userService user.Service, sessionService session.Service, clientService client.Service) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	r.Get("/ping", handler.PingHandler())
	r.Handle("/metrics", promhttp.Handler())

	registerUserRoutes(r, lgr, pr, clientService, userService)
	registerSessionRoutes(r, lgr, pr, clientService, sessionService)
	registerClientRoutes(r, cfg.AuthConfig(), lgr, pr, clientService)

	return r
}

func registerUserRoutes(r chi.Router, lgr *zap.Logger, prometheus reporters.Prometheus, clientService client.Service, userService user.Service) {
	uh := handler.NewUserHandler(userService)

	r.Route("/user", func(r chi.Router) {
		r.Post("/sign-up", withMiddlewares(lgr, prometheus, apiFunc("user", "sign-up"), mdl.WithClientAuth(clientService, mdl.WithError(lgr, uh.SignUp))))
		r.Post("/update-password", withMiddlewares(lgr, prometheus, apiFunc("user", "update-password"), mdl.WithClientAuth(clientService, mdl.WithError(lgr, uh.UpdatePassword))))
	})
}

func registerSessionRoutes(r chi.Router, lgr *zap.Logger, prometheus reporters.Prometheus, clientService client.Service, sessionService session.Service) {
	sh := handler.NewSessionHandler(sessionService)

	r.Route("/session", func(r chi.Router) {
		r.Post("/login", withMiddlewares(lgr, prometheus, apiFunc("session", "login"), mdl.WithClientAuth(clientService, mdl.WithError(lgr, sh.Login))))
		r.Post("/refresh-token", withMiddlewares(lgr, prometheus, apiFunc("session", "refresh-token"), mdl.WithClientAuth(clientService, mdl.WithError(lgr, sh.RefreshToken))))
		r.Post("/logout", withMiddlewares(lgr, prometheus, apiFunc("session", "logout"), mdl.WithClientAuth(clientService, mdl.WithError(lgr, sh.Logout))))
	})
}

func registerClientRoutes(r chi.Router, cfg config.AuthConfig, lgr *zap.Logger, prometheus reporters.Prometheus, clientService client.Service) {
	ch := handler.NewClientHandler(clientService)

	r.Route("/client", func(r chi.Router) {
		r.Use(middleware.BasicAuth("client", map[string]string{cfg.UserName(): cfg.Password()}))
		r.Post("/register", withMiddlewares(lgr, prometheus, apiFunc("client", "create"), mdl.WithError(lgr, ch.Register)))
		r.Post("/revoke", withMiddlewares(lgr, prometheus, apiFunc("client", "revoke"), mdl.WithError(lgr, ch.Revoke)))
	})
}

func apiFunc(api, path string) string {
	return fmt.Sprintf("%s_%s", api, path)
}

//TODO: FIX THE WAY MIDDLEWARE ARE APPLIED
func withMiddlewares(lgr *zap.Logger, prometheus reporters.Prometheus, api string, handler func(resp http.ResponseWriter, req *http.Request), ) http.HandlerFunc {
	return mdl.WithReqRespLog(lgr,
		mdl.WithResponseHeaders(
			mdl.WithPrometheus(prometheus, api, handler),
		),
	)
}
