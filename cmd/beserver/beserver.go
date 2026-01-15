// enc-server-go project main.go
package main

import (
	"flag"
	"log"

	server1 "enc-server-go/pkg/v1-sockets/be/server"
	server2 "enc-server-go/pkg/v2-apis/be/server"
	
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
		log.Fatalf("Failed to start log: %v", err)
	}
	defer logFile.Close()
	log.Println("Started beserver")

	// Load configuration file.
	configs, err := utils.LoadConfigs(configPath)
	if err != nil {
		log.Fatalf("Failed to load configs: %v", err)
	}

	// Verify required configurations.
	if ok, missing := utils.VerifyTopConfigs(configs, []string{"beServerConfigs"}); !ok {
		log.Fatalf("Configuration missing: %s", missing)
	}

	// Make server.
	var s utils.Server
	if v2 {
		log.Println("Running in v2 mode")
		s, err = server2.MakeServer(configs["beServerConfigs"])

	} else {
		log.Println("Running in v1 mode")
		s, err = server1.MakeServer(configs["beServerConfigs"])
	}
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server.
	if err = s.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
