package app

func StartHTTPServer(configFile string) {
	initHTTPServer(configFile).Start()
}
