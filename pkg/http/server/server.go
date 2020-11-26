package server

import (
	"context"
	"identification-service/pkg/config"
	reporters "identification-service/pkg/reporting"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Start()
}

type appServer struct {
	cfg    config.Config
	lgr    reporters.Logger
	router http.Handler
}

func (as *appServer) Start() {
	server := newHTTPServer(as.cfg.HTTPServerConfig(), as.router)

	as.lgr.InfoF("listening on ", as.cfg.HTTPServerConfig().Address())
	go func() { _ = server.ListenAndServe() }()

	waitForShutdown(server, as.lgr)
}

func waitForShutdown(server *http.Server, lgr reporters.Logger) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	defer func() { _ = lgr.Flush() }()

	err := server.Shutdown(context.Background())
	if err != nil {
		lgr.Error(err.Error())
		return
	}

	lgr.Info("server shutdown successful")
}

func newHTTPServer(cfg config.HTTPServerConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         cfg.Address(),
		WriteTimeout: time.Second * time.Duration(cfg.ReadTimeout()),
		ReadTimeout:  time.Second * time.Duration(cfg.WriteTimeout()),
	}
}

func NewServer(cfg config.Config, lgr reporters.Logger, router http.Handler) Server {
	return &appServer{
		cfg:    cfg,
		lgr:    lgr,
		router: router,
	}
}
