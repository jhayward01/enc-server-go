// client
package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"enc-server-go/pkg/utils"
)

// Client implementation.
type clientImpl struct {
	serverAddr string
}


func (c *clientImpl) StoreRecord(id, record []byte) (key []byte, err error) {

	// // Encode data as hex strings.
	// idStr := hex.EncodeToString(id)
	// recordStr := hex.EncodeToString(record)

	// // Write request to server.
	// message, err := c.conn.GetResponse("STORE " + idStr + " " + recordStr + "\n")
	// if err != nil {
	// 	return nil, err
	// }

	// // Check for error.
	// if strings.HasPrefix(message, "ERROR") {
	// 	err = errors.New(message)
	// 	return nil, err
	// }

	// // Decode response
	// if key, err = hex.DecodeString(message); err != nil {
	// 	return nil, err
	// }

	// return key, nil
	
	return nil, nil
}

func (c *clientImpl) RetrieveRecord(id, key []byte) (record []byte, err error) {

	// Example GET request
    getURL := "http://" + c.serverAddr + "/messages"
    resp, err := http.Get(getURL)
    if err != nil {
       	return nil, errors.New("Error making GET request: " + err.Error())
    }
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, errors.New("Error reading response: %v" + err.Error())
        
    }
    fmt.Printf("GET Response:\n%s", string(body))
    
    return body, nil
}

func (c *clientImpl) DeleteRecord(id []byte) (err error) {

	// // Encode data as hex strings
	// idStr := hex.EncodeToString(id)

	// // Write request to server.
	// message, err := c.conn.GetResponse("DELETE " + idStr + "\n")
	// if err != nil {
	// 	return err
	// }

	// // Check for error.
	// if strings.HasPrefix(message, "ERROR") {
	// 	err = errors.New(message)
	// 	return err
	// }

	return nil
}


func MakeClient(configs map[string]string) (c utils.ClientFE, err error) {

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs, []string{"serverAddr"}); !ok {
		err = errors.New("MakeClient missing configuration " + missing)
		return nil, err
	}

	// Build client implementation.
	c = &clientImpl{
		serverAddr: configs["serverAddr"],
	}

	return c, nil
}
