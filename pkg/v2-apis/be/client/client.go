package client

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

// Client implementation.
type clientImpl struct {
	serverAddr string
}

func (c *clientImpl) StoreRecord(id, data []byte) (err error) {

	// Encode data as hex strings
	idStr := hex.EncodeToString(id)
	dataStr := hex.EncodeToString(data)
	
	log.Println("BE client received a store request for", idStr)

	// GRPC connection
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()

	// Create service and context
	s := service.NewBackendServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Process store request
	req := &service.StoreRequest{Id: idStr, Data: dataStr}
	if _, err = s.Store(ctx, req); err != nil {
		return errors.New("Could not send message: " + err.Error())
	}

	return nil
}

func (c *clientImpl) RetrieveRecord(id []byte) (data []byte, err error) {
	
	// Encode data as hex strings
	idStr := hex.EncodeToString(id)
	
	log.Println("BE client received a get request for", idStr)

	// GRPC connection
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()

	// Create service and context
	s := service.NewBackendServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Process get request
	req := &service.RetrieveRequest{Id: idStr}
	resp, err := s.Retrieve(ctx, req)
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
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()

	// Create service and context
	s := service.NewBackendServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Process delete request
	req := &service.DeleteRequest{Id: idStr}
	if _, err = s.Delete(ctx, req); err != nil {
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
	}

	return c, nil
}
