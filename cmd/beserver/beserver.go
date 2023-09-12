// enc-server-go project main.go
package main

import (
	"log"

	"enc-server-go/pkg/be/server"
	"enc-server-go/pkg/utils"
)

const configPath = "config/config.yaml"

func main() {

	// Logging
	logFile, err := utils.StartLog("beserver")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.Println("Started beserver")

	// Load configuration file.
	configs, err := utils.LoadConfigs(configPath)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Make server.
	s, err := server.MakeServer(configs["beServerConfigs"])
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Start server.
	if err = s.Start(); err != nil {
		log.Fatalf(err.Error())
	}
}
