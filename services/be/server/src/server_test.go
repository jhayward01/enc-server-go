package server

import (
	"errors"
	"testing"

	"enc-server-go/utils"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
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
		db: goodDB,
	}

	goodSocketIO, _ = utils.MakeSocketIO(goodServerConfig, goodServer)

	// Cannot set socket IO until server has been created.
	_ = func() bool {
		goodServer.socketIO = goodSocketIO
		return true
	}()

	badDBConfig = map[string]string{
		"foo": "bar"}

	badSocketIOConfig = map[string]string{
		"mongoURI": mongoURI,
	}

	badPortConfig = func() map[string]string {
		m := maps.Clone(goodServerConfig)
		m["port"] = ""
		return m
	}()
)

// Error Descriptions
const badDBMessage = "MakeDB missing configuration mongoURI"
const badSocketIOMessage = "MakeSocketIO missing configuration port"
const badPortMessage = "MakeSocketIO cannot be configured with empty port"

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
			name:    "should fail loading SocketIO",
			args:    args{badSocketIOConfig},
			wantErr: errors.New(badSocketIOMessage),
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

// Respond() - Test Method
func TestServer_Respond(t *testing.T) {

	type fields struct {
		db utils.DB
	}
	type args struct {
		message string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "should run StoreRecord() successfully",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{"STORE " + idHexEncStr + " " + recordHexEncStr},
			want: []byte("SUCCESS\n"),
		},
		{
			name: "should run RetrieveRecord() successfully",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{"RETRIEVE " + idHexEncStr},
			want: []byte(recordHexEncStr + "\n"),
		},
		{
			name: "should run DeleteRecord() successfully",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{"RETRIEVE " + idHexEncStr},
			want: []byte(recordHexEncStr + "\n"),
		},
		{
			name: "should fail on StoreRecord() token count",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{"STORE"},
			want: []byte("ERROR " + badRequest + "\n"),
		},
		{
			name: "should fail on database client StoreRecord()",
			fields: fields{
				db: &MockDB{t, "Store"},
			},
			args: args{"STORE " + idHexEncStr + " " + recordHexEncStr},
			want: []byte("ERROR " + badDBClientMessage + "\n"),
		},
		{
			name: "should fail on RetrieveRecord() token count",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{"STORE"},
			want: []byte("ERROR " + badRequest + "\n"),
		},
		{
			name: "should fail on database client RetrieveRecord()",
			fields: fields{
				db: &MockDB{t, "Retrieve"},
			},
			args: args{"RETRIEVE " + idHexEncStr},
			want: []byte("ERROR " + badDBClientMessage + "\n"),
		},
		{
			name: "should fail on DeleteRecord() token count",
			fields: fields{
				db: &MockDB{t, ""},
			},
			args: args{"DELETE"},
			want: []byte("ERROR " + badRequest + "\n"),
		},
		{
			name: "should fail on database client DeleteRecord()",
			fields: fields{
				db: &MockDB{t, "Delete"},
			},
			args: args{"DELETE " + idHexEncStr},
			want: []byte("ERROR " + badDBClientMessage + "\n"),
		},
	}

	for _, test := range tests {
		s := &serverImpl{
			db:       test.fields.db,
			socketIO: goodSocketIO,
		}

		t.Run(test.name, func(t *testing.T) {
			got := s.Respond(test.args.message)
			assert.Equal(t, test.want, got)
		})
	}
}
