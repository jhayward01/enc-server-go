package service

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dialer interface for dependency injection
type Dialer interface {
	Dial(serverAddr string) (conn *grpc.ClientConn, s BackendServiceClient,
		ctx context.Context, cancel context.CancelFunc, err error)
	Close(conn *grpc.ClientConn, cancel context.CancelFunc)
}

// RealDialer implements the Dialer interface for production use
type RealDialer struct{}

func (d *RealDialer) Dial(serverAddr string) (conn *grpc.ClientConn, s BackendServiceClient,
	ctx context.Context, cancel context.CancelFunc, err error) {

	// GRPC connection
	conn, err = grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, nil, errors.New("Error connecting to backend server: " + err.Error())
	}

	// Create service and context
	s = NewBackendServiceClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

	return conn, s, ctx, cancel, nil
}

func (d *RealDialer) Close(conn *grpc.ClientConn, cancel context.CancelFunc) {
	defer conn.Close()
	defer cancel()
}
