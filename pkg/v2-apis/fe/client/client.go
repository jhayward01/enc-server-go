// client
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
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

	newRecord := record{
		ID:   string(id),
		Data: string(data),
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

	var newRecordWithKey record
	if err = json.Unmarshal(data, &newRecordWithKey); err != nil {
		return nil, errors.New("Error unmarshalling record: " + err.Error())
	}

	return []byte(newRecordWithKey.Key), nil
}

func (c *clientImpl) RetrieveRecord(id, key []byte) (data []byte, err error) {

	getURL := "http://" + c.serverAddr + "/records/" + string(id) + "?key=" + string(key)
	resp, err := http.Get(getURL)
	if err != nil {
		return nil, errors.New("Error making GET request: " + err.Error())
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("Error reading response: " + err.Error())
	}

	return data, nil
}

func (c *clientImpl) DeleteRecord(id []byte) (err error) {

	deleteURL := "http://" + c.serverAddr + "/records/" + string(id)
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

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Error reading response: " + err.Error())
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
