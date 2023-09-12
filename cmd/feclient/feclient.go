package main

import (
	"encoding/hex"
	"log"
	"strconv"

	"enc-server-go/pkg/fe/client"
	"enc-server-go/pkg/utils"
)

const configPath = "config/config.yaml"

func main() {

	// Logging
	logFile, err := utils.StartLog("feclient")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.Println("Started feclient")

	// Load configuration file.
	configs, err := utils.LoadConfigs(configPath)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Set test params
	id := []byte(configs["testParams"]["id"])
	record := []byte(configs["testParams"]["record"])

	// Make client.
	c, err := client.MakeClient(configs["feClientConfigs"])
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Store record.
	log.Println("record", string(record))
	key, err := c.StoreRecord(id, record)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("key", "len="+strconv.Itoa(len(key)), hex.EncodeToString(key))

	// Retrieve record.
	retrieved, err := c.RetrieveRecord(id, key)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("retrieved", string(retrieved))

	// Delete record.
	err = c.DeleteRecord(id)
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("deleted record")
}
