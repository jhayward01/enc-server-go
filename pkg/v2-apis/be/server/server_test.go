package server

import (
	"errors"
	"testing"

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
		serverAddr: "localhost:8888",
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

// Error Descriptions
const badDBMessage = "MakeDB missing configuration mongoURI"
const badPortMessage = "MakeServer missing configuration port"

const badRequest = "Malformed request"
const badDBClientMessage = "Database client error"

// Mock DB Client
type MockDB struct {
	t    *testing.T
	fail string
}

func (db *MockDB) StoreRecord(id, record string) (err error) {
	if db.fail == "Store" {
		return errors.New(badDBClientMessage)
	}
	assert.Equal(db.t, idHexEncStr, id)
	assert.Equal(db.t, recordHexEncStr, record)
	return nil
}

func (db *MockDB) RetrieveRecord(id string) (record string, err error) {
	if db.fail == "Retrieve" {
		return "", errors.New(badDBClientMessage)
	}
	assert.Equal(db.t, idHexEncStr, id)
	return recordHexEncStr, nil
}

func (db *MockDB) DeleteRecord(id string) (err error) {
	if db.fail == "Delete" {
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
				db: &MockDB{t, "Store"},
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
			got, err := s.StoreRecord(nil, test.args.req)
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
				db: &MockDB{t, "Retrieve"},
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
			got, err := s.RetrieveRecord(nil, test.args.req)
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
				db: &MockDB{t, "Delete"},
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
			got, err := s.DeleteRecord(nil, test.args.req)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}
