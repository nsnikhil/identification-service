package main

import "flag"

const (
	configFileKey     = "configFile"
	defaultConfigFile = "local.env"
	configFileUsage   = ""
)

func main() {
	var configFile string
	flag.StringVar(&configFile, configFileKey, defaultConfigFile, configFileUsage)
	flag.Parse()

	execute(flag.Args()[0], configFile)
}
