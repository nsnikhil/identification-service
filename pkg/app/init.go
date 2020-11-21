package app

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"identification-service/pkg/cache"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/http/router"
	"identification-service/pkg/http/server"
	"identification-service/pkg/libcrypto"
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
	dbCfg := cfg.DatabaseConfig()

	sqlDB, err := database.NewHandler(dbCfg).GetDB()
	logError(err)

	db := database.NewSQLDatabase(sqlDB, dbCfg.QueryTTL())

	cc, err := cache.NewHandler(cfg.CacheConfig()).GetCache()
	logError(err)

	//TODO: PASS PROPER LOGGER OR REMOVE LOGGER
	qu := queue.NewQueue(cfg.AMPQConfig().QueueName(), cfg.AMPQConfig().Address(), zap.NewNop())

	en := password.NewEncoder(cfg.PasswordConfig())

	kg := libcrypto.NewKeyGenerator()

	tg, err := token.NewGenerator(cfg.TokenConfig(), kg)
	logError(err)

	cs := initClientService(db, cc, kg)
	us := initUserService(db, en, qu)
	ss := initSessionService(db, us, tg)

	return cs, us, ss
}

func initClientService(db database.SQLDatabase, cc *redis.Client, kg libcrypto.Ed25519Generator) client.Service {
	st := client.NewStore(db, cc)
	return client.NewService(st, kg)
}

func initUserService(db database.SQLDatabase, en password.Encoder, qu queue.Queue) user.Service {
	st := user.NewStore(db)
	return user.NewService(st, en, qu)
}

func initSessionService(db database.SQLDatabase, us user.Service, tg token.Generator) session.Service {
	st := session.NewStore(db)
	return session.NewService(st, us, tg)
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
