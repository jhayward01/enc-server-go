package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"log"
	"strconv"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v1-sockets/fe/client"
)

const configPath = "config/config.yaml"

func main() {

	// Comand line
	var v2 bool
	flag.BoolVar(&v2, "v2", false, "Run in v2 mode")
	flag.Parse()

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
		log.Fatal(err)
	}

	// Verify required configurations.
	if ok, missing := utils.VerifyTopConfigs(configs, []string{"testParams", "feClientConfigs"}); !ok {
		err = errors.New("feclient missing configuration " + missing)
		log.Fatal(err)
	}

	// Set test params
	id := []byte(configs["testParams"]["id"])
	record := []byte(configs["testParams"]["record"])

	// Make client.
	if v2 {
		log.Println("Running in v2 mode")
		
	} else {
		log.Println("Running in v1 mode")
	}
	c, err := client.MakeClient(configs["feClientConfigs"])
	if err != nil {
		log.Fatal(err)
	}

	// Store record.
	log.Println("record", string(record))
	key, err := c.StoreRecord(id, record)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("key", "len="+strconv.Itoa(len(key)), hex.EncodeToString(key))

	// Retrieve record.
	retrieved, err := c.RetrieveRecord(id, key)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("retrieved", string(retrieved))

	// Delete record.
	err = c.DeleteRecord(id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deleted record")
}
