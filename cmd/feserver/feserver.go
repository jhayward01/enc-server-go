// enc-server-go project main.go
package main

import (
	"log"
	"os"

	server1 "enc-server-go/pkg/v1-sockets/fe/server"
	server2 "enc-server-go/pkg/v2-apis/fe/server"
	
	"enc-server-go/pkg/utils"
)

const configPath = "config/config.yaml"

func main() {

	// Comand line
	var v2 bool
	if len(os.Args) > 1 && os.Args[1] == "v2" {
		v2 = true
	}
	
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
	var s utils.Server
	if v2 {
		s, err = server2.MakeServer(configs["feServerConfigs"], configs["beClientConfigs"])
		
	} else {
		s, err = server1.MakeServer(configs["feServerConfigs"], configs["beClientConfigs"])
	}
	
	if err != nil {
		log.Fatal(err.Error())
	}

	if err = s.Start(); err != nil {
		log.Fatal(err.Error())
	}
}
