package app

func StartWorker(configFile string) {
	initConsumer(configFile).Start()
}
