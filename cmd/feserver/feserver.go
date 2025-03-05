// enc-server-go project main.go
package main

import (
	"log"

	"enc-server-go/pkg/fe/server"
	"enc-server-go/pkg/utils"
)

const configPath = "config/config.yaml"

func main() {

	// Logging
	logFile, err := utils.StartLog("feserver")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.Println("Started feserver")

	// Load configuration file.
	configs, err := utils.LoadConfigs(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Make server.
	s, err := server.MakeServer(configs["feServerConfigs"], configs["beClientConfigs"])
	if err != nil {
		log.Fatal(err.Error())
	}

	// Start server.
	if err = s.Start(); err != nil {
		log.Fatal(err.Error())
	}
}
