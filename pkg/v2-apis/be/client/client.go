package client

import (
	"encoding/hex"
	"errors"
	"log"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

// Client implementation.
type clientImpl struct {
	serverAddr string
	dialer     service.Dialer
}

func (c *clientImpl) StoreRecord(id, data []byte) (err error) {

	// Encode data as hex strings
	idStr := hex.EncodeToString(id)
	dataStr := hex.EncodeToString(data)

	log.Println("BE client received a store request for", idStr)

	// GRPC connection
	conn, s, ctx, cancel, err := c.dialer.Dial(c.serverAddr)
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer c.dialer.Close(conn, cancel)

	// Process store request
	req := &service.StoreRequest{Id: idStr, Data: dataStr}
	if _, err = s.StoreRecord(ctx, req); err != nil {
		return errors.New("Could not send message: " + err.Error())
	}

	return nil
}

func (c *clientImpl) RetrieveRecord(id []byte) (data []byte, err error) {

	// Encode data as hex strings
	idStr := hex.EncodeToString(id)

	log.Println("BE client received a get request for", idStr)

	// GRPC connection
	conn, s, ctx, cancel, err := c.dialer.Dial(c.serverAddr)
	if err != nil {
		return nil, errors.New("Error connecting to backend server: " + err.Error())
	}
	defer c.dialer.Close(conn, cancel)

	// Process get request
	req := &service.RetrieveRequest{Id: idStr}
	resp, err := s.RetrieveRecord(ctx, req)
	if err != nil {
		return nil, errors.New("Could not send message: " + err.Error())
	}

	// Decode record from hex.
	if data, err = hex.DecodeString(resp.Data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *clientImpl) DeleteRecord(id []byte) (err error) {

	// Encode data as hex strings
	idStr := hex.EncodeToString(id)

	log.Println("BE client received a delete request for", idStr)

	// GRPC connection
	conn, s, ctx, cancel, err := c.dialer.Dial(c.serverAddr)
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer c.dialer.Close(conn, cancel)

	// Process delete request
	req := &service.DeleteRequest{Id: idStr}
	if _, err = s.DeleteRecord(ctx, req); err != nil {
		return errors.New("Could not send message: " + err.Error())
	}

	return nil
}

func MakeClient(configs map[string]string) (c utils.ClientBE, err error) {

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs, []string{"serverAddr"}); !ok {
		return nil, errors.New("MakeClient missing configuration " + missing)
	}

	// Build client implementation.
	c = &clientImpl{
		serverAddr: configs["serverAddr"],
		dialer: service.Dialer{},
	}

	return c, nil
}
