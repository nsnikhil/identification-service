package app

func StartMigrations(configFile string) {
	logError(initMigrator(configFile).Migrate())
}

func StartRollbacks(configFile string) {
	logError(initMigrator(configFile).Rollback())
}
