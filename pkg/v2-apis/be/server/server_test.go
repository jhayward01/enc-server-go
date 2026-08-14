package server

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

// Test Constants
const port = "8888"
const mongoURI = "mongodb://localhost:27017"

const idHexEncStr = "396263343233393039616335acc30dd405c51d37675d4e0002a526ae113d56"

const recordHexEncStr = "396263343233393039616335b6d61c6839a0dda2524d19b4e5d" +
	"ac5a1fda8902ad2701ced5c31c89088c3151d039ee27d003b75c3a140141c05da496572142eb" +
	"5466c5edb07de33d8ac301f19789fbef68e5c3f280bf4f274e8d2d2d7"

// Error Descriptions
const badDBMessage = "MakeDB missing configuration mongoURI"
const badPortMessage = "MakeServer missing configuration port"

const badRequest = "Malformed request"
const badDBClientMessage = "Database client error"
const badServer = "invalid::address"
const badPort = ":invalid-port"
const badSuccess = "Expected error but server started successfully"

const failedToListenMessage = "Failed to listen"

// MockDB failure modes
const mockDBFailStore = "Store"
const mockDBFailRetrieve = "Retrieve"
const mockDBFailDelete = "Delete"

// Test Variables
var (
	goodDBConfig = map[string]string{
		"port":     port,
		"mongoURI": mongoURI,
	}

	goodDB, _ = utils.MakeDB(goodDBConfig)

	goodServerConfig = map[string]string{
		"port":     port,
		"mongoURI": mongoURI,
	}

	goodServer = &serverImpl{
		db:         goodDB,
		serverAddr: ":8888",
	}

	badDBConfig = func() map[string]string {
		m := maps.Clone(goodServerConfig)
		delete(m, "mongoURI")
		return m
	}()

	badPortConfig = func() map[string]string {
		m := maps.Clone(goodServerConfig)
		delete(m, "port")
		return m
	}()
)

// Mock DB Client
type MockDB struct {
	t    *testing.T
	fail string
}

func (db *MockDB) StoreRecord(id, record string) (err error) {
	if db.fail == mockDBFailStore {
		return errors.New(badDBClientMessage)
	}
	assert.Equal(db.t, idHexEncStr, id)
	assert.Equal(db.t, recordHexEncStr, record)
	return nil
}

func (db *MockDB) RetrieveRecord(id string) (record string, err error) {
	if db.fail == mockDBFailRetrieve {
		return "", errors.New(badDBClientMessage)
	}
	assert.Equal(db.t, idHexEncStr, id)
	return recordHexEncStr, nil
}

func (db *MockDB) DeleteRecord(id string) (err error) {
	if db.fail == mockDBFailDelete {
		return errors.New(badDBClientMessage)
	}
	assert.Equal(db.t, idHexEncStr, id)
	return nil
}

// MakeServer() - Test Method
func TestServer_MakeServer(t *testing.T) {

	type fields struct{}
	type args struct {
		configs map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Server
		wantErr error
	}{
		{
			name: "should run successfully",
			args: args{goodServerConfig},
			want: goodServer,
		},
		{
			name:    "should fail loading DB",
			args:    args{badDBConfig},
			wantErr: errors.New(badDBMessage),
		},
		{
			name:    "should fail on bad port",
			args:    args{badPortConfig},
			wantErr: errors.New(badPortMessage),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeServer(test.args.configs)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// StoreRecord() - Test Method
func TestServer_StoreRecord(t *testing.T) {

	type fields struct {
		db utils.DB
	}
	type args struct {
		req *service.StoreRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.StoreResponse
		wantErr error
	}{
		{
			name: "should run StoreRecord() successfully",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{
				req: &service.StoreRequest{
					Id:   idHexEncStr,
					Data: recordHexEncStr,
				},
			},
			want: &service.StoreResponse{},
		}, {
			name: "should fail on database client StoreRecord()",
			fields: fields{
				db: &MockDB{t, mockDBFailStore},
			},
			args: args{
				req: &service.StoreRequest{
					Id:   idHexEncStr,
					Data: recordHexEncStr,
				},
			},
			wantErr: errors.New(badDBClientMessage),
		},
	}

	for _, test := range tests {
		s := &serverImpl{
			db: test.fields.db,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := s.StoreRecord(context.TODO(), test.args.req)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// RetrieveRecord() - Test Method
func TestServer_RetrieveRecord(t *testing.T) {

	type fields struct {
		db utils.DB
	}
	type args struct {
		req *service.RetrieveRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.RetrieveResponse
		wantErr error
	}{
		{
			name: "should run RetrieveRecord() successfully",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{
				req: &service.RetrieveRequest{
					Id: idHexEncStr,
				},
			},
			want: &service.RetrieveResponse{
				Data: recordHexEncStr,
			},
		}, {
			name: "should fail on database client RetrieveRecord()",
			fields: fields{
				db: &MockDB{t, mockDBFailRetrieve},
			},
			args: args{
				req: &service.RetrieveRequest{
					Id: idHexEncStr,
				},
			},
			wantErr: errors.New(badDBClientMessage),
		},
	}

	for _, test := range tests {
		s := &serverImpl{
			db: test.fields.db,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := s.RetrieveRecord(context.TODO(), test.args.req)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// DeleteRecord() - Test Method
func TestServer_DeleteRecord(t *testing.T) {

	type fields struct {
		db utils.DB
	}
	type args struct {
		req *service.DeleteRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.DeleteResponse
		wantErr error
	}{
		{
			name: "should run DeleteRecord() successfully",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{
				req: &service.DeleteRequest{
					Id: idHexEncStr,
				},
			},
			want: &service.DeleteResponse{},
		}, {
			name: "should fail on database client DeleteRecord()",
			fields: fields{
				db: &MockDB{t, mockDBFailDelete},
			},
			args: args{
				req: &service.DeleteRequest{
					Id: idHexEncStr,
				},
			},
			wantErr: errors.New(badDBClientMessage),
		},
	}

	for _, test := range tests {
		s := &serverImpl{
			db: test.fields.db,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := s.DeleteRecord(context.TODO(), test.args.req)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// Mock Listener for testing Start()
type mockListener struct {
	acceptFn func() (net.Conn, error)
	closeFn  func() error
	addrFn   func() net.Addr
}

func (ml *mockListener) Accept() (net.Conn, error) {
	if ml.acceptFn != nil {
		return ml.acceptFn()
	}
	return nil, errors.New("")
}

func (ml *mockListener) Close() error {
	if ml.closeFn != nil {
		return ml.closeFn()
	}
	return nil
}

func (ml *mockListener) Addr() net.Addr {
	if ml.addrFn != nil {
		return ml.addrFn()
	}
	return nil
}

// Start() - Test Method
func TestServer_Start(t *testing.T) {
	tests := []struct {
		name        string
		serverAddr  string
		wantErr     bool
		errContains string
	}{
		{
			name:        "should fail with invalid address format",
			serverAddr:  badServer,
			wantErr:     true,
			errContains: failedToListenMessage,
		},
		{
			name:        "should fail with invalid port",
			serverAddr:  badPort,
			wantErr:     true,
			errContains: failedToListenMessage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := &serverImpl{
				db:         &MockDB{t, ""},
				serverAddr: test.serverAddr,
			}

			// Start server in a goroutine with a timeout
			done := make(chan error, 1)
			go func() {
				err := server.Start()
				done <- err
			}()

			// Give it a moment to fail
			select {
			case err := <-done:
				if test.wantErr {
					assert.Error(t, err)
					if test.errContains != "" {
						assert.Contains(t, err.Error(), test.errContains)
					}
				} else {
					assert.NoError(t, err)
				}
			case <-time.After(500 * time.Millisecond):
				if test.wantErr {
					t.Error(badSuccess)
				}
			}
		})
	}
}
