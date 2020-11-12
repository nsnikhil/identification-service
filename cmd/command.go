package main

import (
	"identification-service/pkg/app"
	"identification-service/pkg/database"
	"log"
)

const (
	httpServeCommand = "http-serve"
	migrateCommand   = "migrate"
	rollbackCommand  = "rollback"
)

func commands() map[string]func(configFile string) {
	return map[string]func(configFile string){
		httpServeCommand: app.StartHTTPServer,
		migrateCommand:   database.RunMigrations,
		rollbackCommand:  database.RollBackMigrations,
	}
}

func execute(cmd string, configFile string) {
	run, ok := commands()[cmd]
	if !ok {
		log.Fatal("invalid command")
	}

	run(configFile)
}
