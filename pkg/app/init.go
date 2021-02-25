package app

import (
	"database/sql"
	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/nsnikhil/erx"
	"gopkg.in/natefinch/lumberjack.v2"
	"identification-service/pkg/cache"
	"identification-service/pkg/client"
	"identification-service/pkg/config"
	"identification-service/pkg/consumer"
	"identification-service/pkg/database"
	"identification-service/pkg/http/router"
	"identification-service/pkg/http/server"
	"identification-service/pkg/libcrypto"
	"identification-service/pkg/password"
	"identification-service/pkg/producer"
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

func initConsumer(configFile string) consumer.Consumer {
	cfg := config.NewConfig(configFile)
	lgr := initLogger(cfg)
	_, _, ss := initServices(cfg)
	mr := consumer.NewMessageRouter(cfg.KafkaConfig(), ss)
	cs := initKafkaConsumer(initKafkaClient(cfg.KafkaConfig()))

	return consumer.NewConsumer(cfg.KafkaConfig(), lgr, cs, mr)
}

func initKafkaClient(cfg config.KafkaConfig) sarama.Client {
	//TODO: SHOULD COME FORM CONFIGS
	sCfg := sarama.NewConfig()
	sCfg.ClientID = "identification-service"
	sCfg.Producer.Return.Successes = true

	//sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	cl, err := sarama.NewClient(cfg.Addresses(), sCfg)
	logError(err)

	return cl
}

func initKafkaConsumer(cl sarama.Client) sarama.Consumer {
	cs, err := sarama.NewConsumerFromClient(cl)
	logError(err)

	return cs
}

func initKafkaProducer(cl sarama.Client) sarama.SyncProducer {
	sp, err := sarama.NewSyncProducerFromClient(cl)
	logError(err)

	return sp
}

func initReporters(cfg config.Config) (reporters.Logger, reporters.Prometheus) {
	lgr := initLogger(cfg)
	pr := reporters.NewPrometheus()
	return lgr, pr
}

func initMigrator(configFile string) database.Migrator {
	cfg := config.NewConfig(configFile)
	mg, err := database.NewMigrator(config.NewMigrationConfig(), initSqlDB(cfg))
	logError(err)

	return mg
}

func initRouter(cfg config.Config, lgr reporters.Logger, prometheus reporters.Prometheus, cs client.Service, us user.Service, ss session.Service) http.Handler {
	return router.NewRouter(cfg, lgr, prometheus, cs, us, ss)
}

func initSqlDB(cfg config.Config) *sql.DB {
	dbCfg := cfg.DatabaseConfig()

	sqlDB, err := database.NewHandler(dbCfg).GetDB()
	logError(err)

	return sqlDB
}

func initServices(cfg config.Config) (client.Service, user.Service, session.Service) {
	sqlDB := initSqlDB(cfg)

	db := database.NewSQLDatabase(sqlDB, cfg.DatabaseConfig().QueryTTL())

	cc, err := cache.NewHandler(cfg.CacheConfig()).GetCache()
	logError(err)

	kpr := initKafkaProducer(initKafkaClient(cfg.KafkaConfig()))
	pr := producer.NewProducer(kpr)

	en := password.NewEncoder(cfg.PasswordConfig())

	kg := libcrypto.NewKeyGenerator()

	tg, err := token.NewGenerator(cfg.TokenConfig(), kg)
	logError(err)

	cs := initClientService(cfg.ClientConfig(), db, cc, kg)
	us := initUserService(cfg.KafkaConfig(), db, en, pr)
	ss := initSessionService(cfg.ClientConfig(), db, us, tg)

	return cs, us, ss
}

func initClientService(cfg config.ClientConfig, db database.SQLDatabase, cc *redis.Client, kg libcrypto.Ed25519Generator) client.Service {
	st := client.NewStore(db, cc)
	return client.NewService(cfg, st, kg)
}

func initUserService(cfg config.KafkaConfig, db database.SQLDatabase, en password.Encoder, pr producer.Producer) user.Service {
	st := user.NewStore(db)
	return user.NewService(cfg, st, en, pr)
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

	t, ok := err.(*erx.Erx)
	if !ok {
		log.Fatal(err)
	}

	log.Fatal(t.String())
}
