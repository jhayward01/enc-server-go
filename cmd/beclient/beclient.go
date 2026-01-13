// enc-server-go project main.go
package main

import (
	"flag"
	"log"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v1-sockets/be/client"
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
		log.Fatalf("Failed to start log: %v", err)
	}
	defer logFile.Close()
	log.Println("Started beclient")

	// Load configuration file.
	configs, err := utils.LoadConfigs(configPath)
	if err != nil {
		log.Fatalf("Failed to load configs: %v", err)
	}

	// Verify required configurations.
	if ok, missing := utils.VerifyTopConfigs(configs, []string{"testParams", "beClientConfigs"}); !ok {
		log.Fatalf("Configuration missing: %s", missing)
	}

	// Load test params
	id := []byte(configs["testParams"]["id"])
	record := []byte(configs["testParams"]["record"])

	// Make client.
	if v2 {
		log.Println("Running in v2 mode")

	} else {
		log.Println("Running in v1 mode")
	}
	c, err := client.MakeClient(configs["beClientConfigs"])
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Store record.
	log.Println("record", string(record))
	if err = c.StoreRecord(id, record); err != nil {
		log.Fatalf("Failed to store record: %v", err)
	}

	// Retrieve record.
	retrieved, err := c.RetrieveRecord(id)
	if err != nil {
		log.Fatalf("Failed to retrieve record: %v", err)
	}
	log.Println("retrieved", retrieved)

	// Delete record.
	err = c.DeleteRecord(id)
	if err != nil {
		log.Fatalf("Failed to delete record: %v", err)
	}
	log.Println("deleted record")
}
