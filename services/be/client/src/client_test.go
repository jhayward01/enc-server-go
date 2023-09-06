package client

import (
	"errors"
	"testing"

	"enc-server-go/utils"

	"github.com/stretchr/testify/assert"
)

// Test Constants
const serverAddr = "enc-server-go-be:8888"

const idStr = "JTH"
const idHexStr = "4a5448"

const recordStr = "PAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADS"
const recordHexStr = "5041594c4f4144535041594c4f4144535041594c4f414453" +
	"5041594c4f4144535041594c4f4144535041594c4f4144535041594c4f4144535041594" +
	"c4f414453"

const badClientMessage = "MakeClient missing configuration serverAddr"

const storeSuccessMessage = "STORE " + idHexStr + " " + recordHexStr + "\n"
const storeSuccessResponse = ""
const storeFailMessage = "STORE  \n"
const storeFailResponse = "ERROR Malformed request\n"

const retrieveSuccessMessage = "RETRIEVE " + idHexStr + "\n"
const retrieveSuccessResponse = recordHexStr
const retrieveFailMessage = "RETRIEVE \n"
const retrieveFailResponse = "ERROR Malformed request\n"

const deleteSuccessMessage = "DELETE " + idHexStr + "\n"
const deleteSuccessResponse = ""
const deleteFailMessage = "DELETE \n"
const deleteFailResponse = "ERROR Malformed request\n"

// Test Variables
var (
	id     = []byte(idStr)
	record = []byte(recordStr)

	goodClientConfig = map[string]string{
		"serverAddr": serverAddr}

	badClientConfig = map[string]string{
		"foo": "bar"}

	goodConn, _ = utils.MakeConn(goodClientConfig)

	goodClient = &clientImpl{
		conn: goodConn,
	}
)

// Mock Store Connection
type MockConn struct {
	t      *testing.T
	config string
	fail   string
}

func (c MockConn) GetResponse(message string) (response string, err error) {
	switch c.config {
	case "Store":
		if c.fail == "GetResponse" {
			assert.Equal(c.t, storeFailMessage, message)
			return storeFailResponse, nil
		}
		assert.Equal(c.t, storeSuccessMessage, message)
		return storeSuccessResponse, nil

	case "Retrieve":
		if c.fail == "GetResponse" {
			assert.Equal(c.t, retrieveFailMessage, message)
			return retrieveFailResponse, nil
		}
		assert.Equal(c.t, retrieveSuccessMessage, message)
		return retrieveSuccessResponse, nil

	case "Delete":
		if c.fail == "GetResponse" {
			assert.Equal(c.t, deleteFailMessage, message)
			return deleteFailResponse, nil
		}
		assert.Equal(c.t, deleteSuccessMessage, message)
		return deleteSuccessResponse, nil
	}

	return "", nil
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
		want    Client
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

	type fields struct {
		conn utils.Conn
	}
	type args struct {
		id     []byte
		record []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "should run successfully",
			fields: fields{
				conn: MockConn{t, "Store", ""},
			},
			args: args{
				id:     id,
				record: record,
			},
		},
		{
			name: "should return an error",
			fields: fields{
				conn: MockConn{t, "Store", "GetResponse"},
			},
			args: args{
				id:     []byte(""),
				record: []byte(""),
			},
			wantErr: errors.New(storeFailResponse),
		},
	}

	for _, test := range tests {
		c := clientImpl{
			conn: test.fields.conn,
		}

		t.Run(test.name, func(t *testing.T) {
			err := c.StoreRecord(test.args.id, test.args.record)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// RetrieveRecord() - Test Method
func TestClient_RetrieveRecord(t *testing.T) {

	type fields struct {
		conn utils.Conn
	}
	type args struct {
		id []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr error
	}{
		{
			name: "should run successfully",
			fields: fields{
				conn: MockConn{t, "Retrieve", ""},
			},
			args: args{
				id: id,
			},
			want: record,
		},
		{
			name: "should return an error",
			fields: fields{
				conn: MockConn{t, "Retrieve", "GetResponse"},
			},
			args: args{
				id: []byte(""),
			},
			wantErr: errors.New(retrieveFailResponse),
		},
	}

	for _, test := range tests {
		c := clientImpl{
			conn: test.fields.conn,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := c.RetrieveRecord(test.args.id)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// DeleteRecord() - Test Method
func TestClient_DeleteRecord(t *testing.T) {

	type fields struct {
		conn utils.Conn
	}
	type args struct {
		id []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "should run successfully",
			fields: fields{
				conn: MockConn{t, "Delete", ""},
			},
			args: args{
				id: id,
			},
		},
		{
			name: "should return an error",
			fields: fields{
				conn: MockConn{t, "Delete", "GetResponse"},
			},
			args: args{
				id: []byte(""),
			},
			wantErr: errors.New(retrieveFailResponse),
		},
	}

	for _, test := range tests {
		c := clientImpl{
			conn: test.fields.conn,
		}

		t.Run(test.name, func(t *testing.T) {
			err := c.DeleteRecord(test.args.id)
			assert.Equal(t, test.wantErr, err)
		})
	}
}
