// enc-server-go project main.go
package main

import (
	"flag"
	"log"

	"enc-server-go/pkg/v1-sockets/be/client"
	"enc-server-go/pkg/utils"
)

const configPath = "config/config.yaml"

func main() {

	// Comand line
	var v2 bool
	flag.BoolVar(&v2, "v2", false, "Run in v2 mode")
	flag.Parse()

	// Logging
	logFile, err := utils.StartLog("beclient")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.Println("Started beclient")

	// Load configuration file.
	configs, err := utils.LoadConfigs(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Load test params
	id := []byte(configs["testParams"]["id"])
	record := []byte(configs["testParams"]["record"])

	// Make client.
	c, err := client.MakeClient(configs["beClientConfigs"])
	if err != nil {
		log.Fatal(err)
	}

	// Store record.
	log.Println("record", string(record))
	if err = c.StoreRecord(id, record); err != nil {
		log.Fatal(err)
	}

	// Retrieve record.
	retrieved, err := c.RetrieveRecord(id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("retrieved", retrieved)

	// Delete record.
	err = c.DeleteRecord(id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deleted record")
}
