// client
package client

import (
	"encoding/hex"
	"errors"
	"strings"

	"enc-server-go/pkg/utils"
)

// Client implementation.
type clientImpl struct {
	conn utils.Conn
}

func (c *clientImpl) StoreRecord(id, record []byte) (key []byte, err error) {

	// Encode data as hex strings.
	idStr := hex.EncodeToString(id)
	recordStr := hex.EncodeToString(record)

	// Write request to server.
	message, err := c.conn.GetResponse("STORE " + idStr + " " + recordStr + "\n")
	if err != nil {
		return nil, err
	}

	// Check for error.
	if strings.HasPrefix(message, "ERROR") {
		err = errors.New(message)
		return nil, err
	}

	// Decode response
	if key, err = hex.DecodeString(message); err != nil {
		return nil, err
	}

	return key, nil
}

func (c *clientImpl) RetrieveRecord(id, key []byte) (record []byte, err error) {

	// Encode data as hex strings
	idStr := hex.EncodeToString(id)
	keyStr := hex.EncodeToString(key)

	// Write request to server.
	message, err := c.conn.GetResponse("RETRIEVE " + idStr + " " + keyStr + "\n")
	if err != nil {
		return nil, err
	}

	// Check for error.
	if strings.HasPrefix(message, "ERROR") {
		err = errors.New(message)
		return nil, err
	}

	// Decode response
	if record, err = hex.DecodeString(message); err != nil {
		return nil, err
	}

	return record, nil
}

func (c *clientImpl) DeleteRecord(id []byte) (err error) {

	// Encode data as hex strings
	idStr := hex.EncodeToString(id)

	// Write request to server.
	message, err := c.conn.GetResponse("DELETE " + idStr + "\n")
	if err != nil {
		return err
	}

	// Check for error.
	if strings.HasPrefix(message, "ERROR") {
		err = errors.New(message)
		return err
	}

	return nil
}

func MakeClient(configs map[string]string) (c utils.ClientFE, err error) {

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs, []string{"serverAddr"}); !ok {
		err = errors.New("MakeClient missing configuration " + missing)
		return nil, err
	}

	// Build connection object.
	conn, err := utils.MakeConn(configs)
	if err != nil {
		return nil, err
	}

	// Build client implementation.
	c = &clientImpl{
		conn: conn,
	}

	return c, nil
}
