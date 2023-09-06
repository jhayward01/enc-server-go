// enc-server-go project main.go
package main

import (
	"log"

	"enc-server-go/services/be/client/src"
	"enc-server-go/utils"
)

const configPath = "config/config.yaml"

func main() {

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
		log.Fatalf(err.Error())
	}

	// Load test params
	id := []byte(configs["testParams"]["id"])
	record := []byte(configs["testParams"]["record"])

	// Make client.
	c, err := client.MakeClient(configs["beClientConfigs"])
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Store record.
	log.Println("record", string(record))
	if err = c.StoreRecord(id, record); err != nil {
		log.Fatalf(err.Error())
	}

	// Retrieve record.
	retrieved, err := c.RetrieveRecord(id)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("retrieved", retrieved)

	// Delete record.
	err = c.DeleteRecord(id)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("deleted record")
}
