// enc-server-go project main.go
package main

import (
	"errors"
	"flag"
	"log"

	"enc-server-go/pkg/v1-sockets/be/server"
	"enc-server-go/pkg/utils"
)

const configPath = "config/config.yaml"

func main() {

	// Comand line
	var v2 bool
	flag.BoolVar(&v2, "v2", false, "Run in v2 mode")
	flag.Parse()

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
		log.Fatal(err)
	}

	// Verify required configurations.
	if ok, missing := utils.VerifyTopConfigs(configs, []string{"beServerConfigs"}); !ok {
		err = errors.New("feserver missing configuration " + missing)
		log.Fatal(err)
	}

	// Make server.
	if v2 {
		log.Println("Running in v2 mode")
		
	} else {
		log.Println("Running in v1 mode")
	}
	s, err := server.MakeServer(configs["beServerConfigs"])
	if err != nil {
		log.Fatal(err)
	}

	// Start server.
	if err = s.Start(); err != nil {
		log.Fatal(err)
	}
}
