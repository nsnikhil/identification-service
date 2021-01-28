package main

import (
	"identification-service/pkg/app"
	"log"
)

const (
	httpServeCommand = "http-serve"
	workerCommand    = "worker"
	migrateCommand   = "migrate"
	rollbackCommand  = "rollback"
)

func commands() map[string]func(configFile string) {
	return map[string]func(configFile string){
		httpServeCommand: app.StartHTTPServer,
		workerCommand:    app.StartWorker,
		migrateCommand:   app.StartMigrations,
		rollbackCommand:  app.StartRollbacks,
	}
}

func execute(cmd string, configFile string) {
	run, ok := commands()[cmd]
	if !ok {
		log.Fatal("invalid command")
	}

	run(configFile)
}
