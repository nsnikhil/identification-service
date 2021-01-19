package app

import (
	"github.com/go-redis/redis/v8"
	"gopkg.in/natefinch/lumberjack.v2"
	"identification-service/pkg/cache"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/database"
	"identification-service/pkg/event/publisher"
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

func initReporters(cfg config.Config) (reporters.Logger, reporters.Prometheus) {
	lgr := initLogger(cfg)
	pr := reporters.NewPrometheus()
	return lgr, pr
}

func initRouter(cfg config.Config, lgr reporters.Logger, prometheus reporters.Prometheus, cs client.Service, us user.Service, ss session.Service) http.Handler {
	return router.NewRouter(cfg, lgr, prometheus, cs, us, ss)
}

func initServices(cfg config.Config) (client.Service, user.Service, session.Service) {
	dbCfg := cfg.DatabaseConfig()

	sqlDB, err := database.NewHandler(dbCfg).GetDB()
	logError(err)

	db := database.NewSQLDatabase(sqlDB, dbCfg.QueryTTL())

	cc, err := cache.NewHandler(cfg.CacheConfig()).GetCache()
	logError(err)

	qu := queue.NewAMQP(cfg.AMPQConfig().Address())

	pr, err := publisher.NewPublisher(qu, cfg.PublisherConfig().QueueMap())
	logError(err)

	en := password.NewEncoder(cfg.PasswordConfig())

	kg := libcrypto.NewKeyGenerator()

	tg, err := token.NewGenerator(cfg.TokenConfig(), kg)
	logError(err)

	cs := initClientService(cfg.ClientConfig(), db, cc, kg)
	us := initUserService(db, en, pr)
	ss := initSessionService(cfg.ClientConfig(), db, us, tg)

	return cs, us, ss
}

func initClientService(cfg config.ClientConfig, db database.SQLDatabase, cc *redis.Client, kg libcrypto.Ed25519Generator) client.Service {
	st := client.NewStore(db, cc)
	return client.NewService(cfg, st, kg)
}

func initUserService(db database.SQLDatabase, en password.Encoder, pr publisher.Publisher) user.Service {
	st := user.NewStore(db)
	return user.NewService(st, en, pr)
}

func initSessionService(cfg config.ClientConfig, db database.SQLDatabase, us user.Service, tg token.Generator) session.Service {
	st := session.NewStore(db)
	sts := initStrategies(cfg, st)
	return session.NewService(st, us, tg, sts)
}

//TODO: NAME SHOULD COME FROM CONFIG
func initStrategies(cfg config.ClientConfig, store session.Store) map[string]session.Strategy {
	res := make(map[string]session.Strategy)

	for strategy := range cfg.Strategies() {
		switch strategy {
		case "revoke_old":
			res[strategy] = session.NewRevokeOldStrategy(store)
		}
	}

	return res
}

func initLogger(cfg config.Config) reporters.Logger {
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
		"file":   newExternalLogFile(cfg.LogFileConfig()),
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

func newExternalLogFile(cfg config.LogFileConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   cfg.GetFilePath(),
		MaxSize:    cfg.GetFileMaxSizeInMb(),
		MaxBackups: cfg.GetFileMaxBackups(),
		MaxAge:     cfg.GetFileMaxAge(),
		LocalTime:  cfg.GetFileWithLocalTimeStamp(),
	}
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
