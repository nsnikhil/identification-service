package app

import (
	"go.uber.org/zap"
	"identification-service/pkg/cache"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/http/router"
	"identification-service/pkg/http/server"
	"identification-service/pkg/liberr"
	"identification-service/pkg/password"
	"identification-service/pkg/queue"
	reporters "identification-service/pkg/reporting"
	"identification-service/pkg/session"
	"identification-service/pkg/token"
	"identification-service/pkg/user"
	"io"
	"log"
	"net/http"
	"os"
)

func initHTTPServer(configFile string) server.Server {
	cfg := config.NewConfig(configFile)
	lgr, pr := initReporters(cfg)
	cs, us, ss := initServices(cfg)
	rt := initRouter(cfg, lgr, pr, cs, us, ss)
	return server.NewServer(cfg, lgr, rt)
}

func initReporters(cfg config.Config) (*zap.Logger, reporters.Prometheus) {
	lgr := initLogger(cfg)
	pr := reporters.NewPrometheus()
	return lgr, pr
}

func initRouter(cfg config.Config, lgr *zap.Logger, prometheus reporters.Prometheus, cs client.Service, us user.Service, ss session.Service) http.Handler {
	return router.NewRouter(cfg, lgr, prometheus, cs, us, ss)
}

func initServices(cfg config.Config) (client.Service, user.Service, session.Service) {
	db, err := database.NewHandler(cfg.DatabaseConfig()).GetDB()
	logError(err)

	cc, err := cache.NewHandler(cfg.CacheConfig()).GetCache()
	logError(err)

	//TODO: PASS PROPER LOGGER OR REMOVE LOGGER
	qu := queue.NewQueue(cfg.AMPQConfig().QueueName(), cfg.AMPQConfig().Address(), zap.NewNop())

	ec := password.NewEncoder(cfg.PasswordConfig())

	gn, err := token.NewGenerator(cfg.TokenConfig())
	logError(err)

	cs := client.NewService(db, cc)
	us := user.NewService(db, ec, qu)
	ss := session.NewService(db, us, cs, gn)

	return cs, us, ss
}

func initLogger(cfg config.Config) *zap.Logger {
	return reporters.NewLogger(
		cfg.Env(),
		cfg.LogConfig().Level(),
		getWriters(cfg)...,
	)
}

func getWriters(cfg config.Config) []io.Writer {
	//TODO: MOVE TO CONST
	logSinkMap := map[string]io.Writer{
		"stdout": os.Stdout,
		"file":   reporters.NewExternalLogFile(cfg.LogFileConfig()),
	}

	var writers []io.Writer
	for _, sink := range cfg.LogConfig().Sinks() {
		w, ok := logSinkMap[sink]
		if ok {
			writers = append(writers, w)
		}
	}

	return writers
}

func logError(err error) {
	if err == nil {
		return
	}

	t, ok := err.(*liberr.Error)
	if !ok {
		log.Fatal(err)
	}

	log.Fatal(t.EncodedStack())
}
