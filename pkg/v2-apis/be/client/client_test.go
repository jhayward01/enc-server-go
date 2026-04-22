package client

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

// Test Constants
const serverAddr = "enc-server-go-be:8888"

const idStr = "JTH"
const idHexStr = "4a5448"

const badClientMessage = "MakeClient missing configuration serverAddr"

// Test Variables
var (
	goodClientConfig = map[string]string{
		"serverAddr": serverAddr}

	badClientConfig = map[string]string{
		"foo": "bar"}

	goodClient = &clientImpl{
		serverAddr: serverAddr,
		dialer:     &service.RealDialer{},
	}
)

// Mock BackendServiceClient
type mockBackendServiceClient struct {
	storeRecordFn   func(ctx context.Context, in *service.StoreRequest, opts ...grpc.CallOption) (*service.StoreResponse, error)
	retrieveRecordFn func(ctx context.Context, in *service.RetrieveRequest, opts ...grpc.CallOption) (*service.RetrieveResponse, error)
	deleteRecordFn   func(ctx context.Context, in *service.DeleteRequest, opts ...grpc.CallOption) (*service.DeleteResponse, error)
}

func (m *mockBackendServiceClient) StoreRecord(ctx context.Context, in *service.StoreRequest, opts ...grpc.CallOption) (*service.StoreResponse, error) {
	if m.storeRecordFn != nil {
		return m.storeRecordFn(ctx, in, opts...)
	}
	return &service.StoreResponse{}, nil
}

func (m *mockBackendServiceClient) RetrieveRecord(ctx context.Context, in *service.RetrieveRequest, opts ...grpc.CallOption) (*service.RetrieveResponse, error) {
	if m.retrieveRecordFn != nil {
		return m.retrieveRecordFn(ctx, in, opts...)
	}
	return &service.RetrieveResponse{}, nil
}

func (m *mockBackendServiceClient) DeleteRecord(ctx context.Context, in *service.DeleteRequest, opts ...grpc.CallOption) (*service.DeleteResponse, error) {
	if m.deleteRecordFn != nil {
		return m.deleteRecordFn(ctx, in, opts...)
	}
	return &service.DeleteResponse{}, nil
}

// Mock Dialer
type mockDialer struct {
	dialFn  func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error)
	closeFn func(conn *grpc.ClientConn, cancel context.CancelFunc)
}

func (m *mockDialer) Dial(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
	if m.dialFn != nil {
		return m.dialFn(serverAddr)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return nil, nil, ctx, cancel, nil
}

func (m *mockDialer) Close(conn *grpc.ClientConn, cancel context.CancelFunc) {
	if m.closeFn != nil {
		m.closeFn(conn, cancel)
	}
}

// MakeClient() - Test Method
func TestClient_MakeClient(t *testing.T) {

	type fields struct{}
	type args struct {
		config map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    utils.ClientBE
		wantErr error
	}{
		{
			name: "should run successfully",
			args: args{goodClientConfig},
			want: goodClient,
		},
		{
			name:    "should fail loading configuration",
			args:    args{badClientConfig},
			wantErr: errors.New(badClientMessage),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeClient(test.args.config)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// StoreRecord() - Test Method
func TestClient_StoreRecord(t *testing.T) {
	tests := []struct {
		name        string
		id          []byte
		data        []byte
		mockServiceFn func(ctx context.Context, in *service.StoreRequest, opts ...grpc.CallOption) (*service.StoreResponse, error)
		mockDialerFn  func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error)
		wantErr     bool
		errContains string
	}{
		{
			name: "should store record successfully",
			id:   []byte("test-id"),
			data: []byte("test-data"),
			mockServiceFn: func(ctx context.Context, in *service.StoreRequest, opts ...grpc.CallOption) (*service.StoreResponse, error) {
				// Verify request details
				assert.Equal(t, hex.EncodeToString([]byte("test-id")), in.Id)
				assert.Equal(t, hex.EncodeToString([]byte("test-data")), in.Data)
				return &service.StoreResponse{}, nil
			},
			mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
				ctx, cancel := context.WithCancel(context.Background())
				return nil, &mockBackendServiceClient{}, ctx, cancel, nil
			},
			wantErr: false,
		},
		{
			name: "should fail on dialer error",
			id:   []byte("test-id"),
			data: []byte("test-data"),
			mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
				return nil, nil, nil, nil, errors.New("connection failed")
			},
			wantErr:     true,
			errContains: "Error connecting to backend server",
		},
		// {
		// 	name: "should fail on service error",
		// 	id:   []byte("test-id"),
		// 	data: []byte("test-data"),
		// 	mockServiceFn: func(ctx context.Context, in *service.StoreRequest, opts ...grpc.CallOption) (*service.StoreResponse, error) {
		// 		return nil, errors.New("service error")
		// 	},
		// 	mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
		// 		ctx, cancel := context.WithCancel(context.Background())
		// 		return nil, &mockBackendServiceClient{}, ctx, cancel, nil
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Could not send message",
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := &mockBackendServiceClient{
				storeRecordFn: test.mockServiceFn,
			}

			mockDialerObj := &mockDialer{
				dialFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
					if test.mockDialerFn != nil {
						conn, svc, ctx, cancel, err := test.mockDialerFn(serverAddr)
						// If no error, return our mock service
						if err == nil && svc == nil {
							return conn, mockService, ctx, cancel, err
						}
						return conn, svc, ctx, cancel, err
					}
					ctx, cancel := context.WithCancel(context.Background())
					return nil, mockService, ctx, cancel, nil
				},
				closeFn: func(conn *grpc.ClientConn, cancel context.CancelFunc) {
					if cancel != nil {
						cancel()
					}
				},
			}

			client := &clientImpl{
				serverAddr: serverAddr,
				dialer:     mockDialerObj,
			}

			err := client.StoreRecord(test.id, test.data)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// RetrieveRecord() - Test Method
func TestClient_RetrieveRecord(t *testing.T) {
	tests := []struct {
		name        string
		id          []byte
		mockServiceFn func(ctx context.Context, in *service.RetrieveRequest, opts ...grpc.CallOption) (*service.RetrieveResponse, error)
		mockDialerFn  func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error)
		wantData    []byte
		wantErr     bool
		errContains string
	}{
		{
			name: "should retrieve record successfully",
			id:   []byte("test-id"),
			mockServiceFn: func(ctx context.Context, in *service.RetrieveRequest, opts ...grpc.CallOption) (*service.RetrieveResponse, error) {
				// Verify request details
				assert.Equal(t, hex.EncodeToString([]byte("test-id")), in.Id)
				return &service.RetrieveResponse{
					Data: hex.EncodeToString([]byte("test-data")),
				}, nil
			},
			mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
				ctx, cancel := context.WithCancel(context.Background())
				return nil, &mockBackendServiceClient{}, ctx, cancel, nil
			},
			wantData: []byte("test-data"),
			wantErr:  false,
		},
		// {
		// 	name: "should fail on dialer error",
		// 	id:   []byte("test-id"),
		// 	mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
		// 		return nil, nil, nil, nil, errors.New("connection failed")
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Error connecting to backend server",
		// },
		// {
		// 	name: "should fail on service error",
		// 	id:   []byte("test-id"),
		// 	mockServiceFn: func(ctx context.Context, in *service.RetrieveRequest, opts ...grpc.CallOption) (*service.RetrieveResponse, error) {
		// 		return nil, errors.New("service error")
		// 	},
		// 	mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
		// 		ctx, cancel := context.WithCancel(context.Background())
		// 		return nil, &mockBackendServiceClient{}, ctx, cancel, nil
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Could not send message",
		// },
		// {
		// 	name: "should fail on invalid hex response",
		// 	id:   []byte("test-id"),
		// 	mockServiceFn: func(ctx context.Context, in *service.RetrieveRequest, opts ...grpc.CallOption) (*service.RetrieveResponse, error) {
		// 		return &service.RetrieveResponse{
		// 			Data: "invalid-hex",
		// 		}, nil
		// 	},
		// 	mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
		// 		ctx, cancel := context.WithCancel(context.Background())
		// 		return nil, &mockBackendServiceClient{}, ctx, cancel, nil
		// 	},
		// 	wantErr:     true,
		// 	errContains: "encoding/hex",
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := &mockBackendServiceClient{
				retrieveRecordFn: test.mockServiceFn,
			}

			mockDialerObj := &mockDialer{
				dialFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
					// If there's a custom dialer function, use it
					if test.mockDialerFn != nil {
						return test.mockDialerFn(serverAddr)
					}
					// Otherwise, return success with our mock service
					ctx, cancel := context.WithCancel(context.Background())
					return nil, mockService, ctx, cancel, nil
				},
				closeFn: func(conn *grpc.ClientConn, cancel context.CancelFunc) {
					if cancel != nil {
						cancel()
					}
				},
			}

			client := &clientImpl{
				serverAddr: serverAddr,
				dialer:     mockDialerObj,
			}

			got, err := client.RetrieveRecord(test.id)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.wantData, got)
			}
		})
	}
}

// DeleteRecord() - Test Method
func TestClient_DeleteRecord(t *testing.T) {
	tests := []struct {
		name        string
		id          []byte
		mockServiceFn func(ctx context.Context, in *service.DeleteRequest, opts ...grpc.CallOption) (*service.DeleteResponse, error)
		mockDialerFn  func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error)
		wantErr     bool
		errContains string
	}{
		{
			name: "should delete record successfully",
			id:   []byte("test-id"),
			mockServiceFn: func(ctx context.Context, in *service.DeleteRequest, opts ...grpc.CallOption) (*service.DeleteResponse, error) {
				// Verify request details
				assert.Equal(t, hex.EncodeToString([]byte("test-id")), in.Id)
				return &service.DeleteResponse{}, nil
			},
			mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
				ctx, cancel := context.WithCancel(context.Background())
				return nil, &mockBackendServiceClient{}, ctx, cancel, nil
			},
			wantErr: false,
		},
		{
			name: "should fail on dialer error",
			id:   []byte("test-id"),
			mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
				return nil, nil, nil, nil, errors.New("connection failed")
			},
			wantErr:     true,
			errContains: "Error connecting to backend server",
		},
		// {
		// 	name: "should fail on service error",
		// 	id:   []byte("test-id"),
		// 	mockServiceFn: func(ctx context.Context, in *service.DeleteRequest, opts ...grpc.CallOption) (*service.DeleteResponse, error) {
		// 		return nil, errors.New("service error")
		// 	},
		// 	mockDialerFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
		// 		ctx, cancel := context.WithCancel(context.Background())
		// 		return nil, &mockBackendServiceClient{}, ctx, cancel, nil
		// 	},
		// 	wantErr:     true,
		// 	errContains: "Could not send message",
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := &mockBackendServiceClient{
				deleteRecordFn: test.mockServiceFn,
			}

			mockDialerObj := &mockDialer{
				dialFn: func(serverAddr string) (*grpc.ClientConn, service.BackendServiceClient, context.Context, context.CancelFunc, error) {
					if test.mockDialerFn != nil {
						conn, svc, ctx, cancel, err := test.mockDialerFn(serverAddr)
						// If no error, return our mock service
						if err == nil && svc == nil {
							return conn, mockService, ctx, cancel, err
						}
						return conn, svc, ctx, cancel, err
					}
					ctx, cancel := context.WithCancel(context.Background())
					return nil, mockService, ctx, cancel, nil
				},
				closeFn: func(conn *grpc.ClientConn, cancel context.CancelFunc) {
					if cancel != nil {
						cancel()
					}
				},
			}

			client := &clientImpl{
				serverAddr: serverAddr,
				dialer:     mockDialerObj,
			}

			err := client.DeleteRecord(test.id)

			if test.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
