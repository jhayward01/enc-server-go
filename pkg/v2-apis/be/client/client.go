package client

import (
	"context"
	"errors"
	"time"
	
	"google.golang.org/grpc"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

// Client implementation.
type clientImpl struct {
	serverAddr string
}

func (c *clientImpl) StoreRecord(id, record []byte) (err error) {
	
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()
	
	s := service.NewBackendServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &service.StoreRequest{Id: string(id), Data: string(record)}
	if _, err = s.Store(ctx, req); err != nil {
		return errors.New("Could not send message: " + err.Error())
	}

	return nil
}

func (c *clientImpl) RetrieveRecord(id []byte) (record []byte, err error) {
	
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()
	
	s := service.NewBackendServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	req := &service.RetrieveRequest{Id: string(id)}
	resp, err := s.Retrieve(ctx, req)
	if err != nil {
		return nil, errors.New("Could not send message: " + err.Error())
	}
	
	return []byte(resp.Data), nil
}

func (c *clientImpl) DeleteRecord(id []byte) (err error) {
	
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()
	
	s := service.NewBackendServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	req := &service.DeleteRequest{Id: string(id)}
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
