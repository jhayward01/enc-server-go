// client
package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"enc-server-go/pkg/utils"
)

type record struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Data string `json:"data"`
}

// Client implementation.
type clientImpl struct {
	serverAddr string
}

func (c *clientImpl) StoreRecord(id, data []byte) (key []byte, err error) {

	// Encode data as hex strings.
	idStr := hex.EncodeToString(id)
	dataStr := hex.EncodeToString(data)

	log.Println("FE client received a store request for", idStr, dataStr)

	newRecord := record{
		ID:   idStr,
		Data: dataStr,
	}

	jsonData, err := json.Marshal(newRecord)
	if err != nil {
		return nil, errors.New("Error marshaling JSON: " + err.Error())
	}

	postURL := "http://" + c.serverAddr + "/records"
	resp, err := http.Post(postURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.New("Error making POST request: " + err.Error())
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error reading response: " + err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New("Bad status making POST request: " + resp.Status + " " + string(data))
	}

	if err = json.Unmarshal(data, &newRecord); err != nil {
		return nil, errors.New("Error unmarshalling record: " + err.Error())
	}

	if key, err = hex.DecodeString(newRecord.Key); err != nil {
		return nil, errors.New("Error decoding key: " + err.Error())
	}
	return key, nil
}

func (c *clientImpl) RetrieveRecord(id, key []byte) (data []byte, err error) {

	// Encode data as hex strings.
	idStr := hex.EncodeToString(id)
	keyStr := hex.EncodeToString(key)

	log.Println("FE client received a retrieve request for", idStr, keyStr)

	getURL := "http://" + c.serverAddr + "/records/" + idStr + "?key=" + keyStr
	resp, err := http.Get(getURL)
	if err != nil {
		return nil, errors.New("Error making GET request: " + err.Error())
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error reading response: " + err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Bad status making GET request: " + resp.Status + " " + string(data))
	}

	return data, nil
}

func (c *clientImpl) DeleteRecord(id []byte) (err error) {

	// Encode data as hex strings.
	idStr := hex.EncodeToString(id)

	log.Println("FE client received a delete request for", idStr)

	deleteURL := "http://" + c.serverAddr + "/records/" + idStr
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return errors.New("Error composing DELETE request: " + err.Error())
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Error making DELETE request: " + err.Error())
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Error reading response: " + err.Error())
	}

	if resp.StatusCode != http.StatusAccepted {
		return errors.New("Bad status making DELETE request: " + resp.Status + " " + string(data))
	}

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
