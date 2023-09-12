package server

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"enc-server-go/pkg/be/client"
	"enc-server-go/pkg/utils"
)

// Test Constants
const idKeyStr = "vkAZAarLbZ6w0kmL2HJP3eU1ODCgVj4k"
const idNonceStr = "9bc423909ac5"
const keySizeStr = "32"
const port = "7777"

const serverAddr = "enc-server-go-be:8888"

const badKeySizeStr = "fff"

const idStr = "JTH"
const idHexStr = "4a5448"
const idHexEncStr = "396263343233393039616335acc30dd405c51d37675d4e0002a526ae113d56"

const recordStr = "PAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADS"
const recordHexStr = "5041594c4f4144535041594c4f4144535041594c4f414453504159" +
	"4c4f4144535041594c4f4144535041594c4f4144535041594c4f4144535041594c4f414453"
const recordHexEncStr = "396263343233393039616335b6d61c6839a0dda2524d19b4e5d" +
	"ac5a1fda8902ad2701ced5c31c89088c3151d039ee27d003b75c3a140141c05da496572142eb" +
	"5466c5edb07de33d8ac301f19789fbef68e5c3f280bf4f274e8d2d2d7"

// Test Variables
var (
	id     = []byte(idStr)
	record = []byte(recordStr)

	idEnc, _ = hex.DecodeString(idHexEncStr)

	recordEnc, _ = hex.DecodeString(recordHexEncStr)

	idKey   = []byte(idKeyStr)
	idNonce = []byte(idNonceStr)

	idKeyHexStr = hex.EncodeToString(idKey)

	keygen, _ = utils.MakeKeyGen(map[string]string{"keySize": keySizeStr})

	idCipher, _ = keygen.GetGCMCipher([]byte(idKeyStr))

	goodClientConfig = map[string]string{
		"serverAddr": serverAddr}

	goodClient, _ = client.MakeClient(goodClientConfig)

	goodServerConfig = map[string]string{
		"idKeyStr":   idKeyStr,
		"idNonceStr": idNonceStr,
		"keySize":    keySizeStr,
		"port":       port,
	}

	goodServer = &serverImpl{
		keygen:   keygen,
		idNonce:  idNonce,
		idCipher: idCipher,
		beClient: goodClient,
	}

	goodSocketIO, _ = utils.MakeSocketIO(goodServerConfig, goodServer)

	// Cannot set socket IO until server has been created.
	_ = func() bool {
		goodServer.socketIO = goodSocketIO
		return true
	}()

	badServerConfig = map[string]string{
		"foo": "bar"}

	badKeyGenConfig = func() map[string]string {
		m := maps.Clone(goodServerConfig)
		m["keySize"] = badKeySizeStr
		return m
	}()

	badIdKeyConfig = func() map[string]string {
		m := maps.Clone(goodServerConfig)
		m["idKeyStr"] = ""
		return m
	}()

	badClientConfig = map[string]string{
		"foo": "bar"}

	badPortConfig = func() map[string]string {
		m := maps.Clone(goodServerConfig)
		m["port"] = ""
		return m
	}()
)

// Error Descriptions
const badServerMessage = "MakeServer missing configuration keySize"
const badClientMessage = "MakeClient missing configuration serverAddr"
const badSocketIOMessage = "MakeSocketIO cannot be configured with empty port"
const badRandomKeyMessage = "KeyGen.RandomKey error"
const badGetGCMCipherMessage = "KeyGen.GetGCMKey error"
const badRandomNonceMessage = "KeyGen.RandomNonce error"
const badBEClientMessage = "Back-end client error"
const badDecryptMessage = "cipher: message authentication failed"
const badRequest = "Malformed request"
const badDecode = "encoding/hex: invalid byte: U+0067 'g'"

// Mock KeyGen
type MockKeyGen struct {
	t    *testing.T
	fail string
}

func (k *MockKeyGen) RandomKey() (key []byte, err error) {
	if k.fail == "RandomKey" {
		return nil, errors.New(badRandomKeyMessage)
	}
	return idKey, nil
}

func (k *MockKeyGen) GetGCMCipher(key []byte) (gcmCipher cipher.AEAD, err error) {
	if k.fail == "GetGCMCipher" {
		return nil, errors.New(badGetGCMCipherMessage)
	}
	keygen, _ := utils.MakeKeyGen(goodServerConfig)
	return keygen.GetGCMCipher(key)
}

func (k *MockKeyGen) RandomNonce(nonceSize int) (nonce []byte, err error) {
	if k.fail == "RandomNonce" {
		return nil, errors.New(badRandomNonceMessage)
	}
	return idNonce, nil
}

// Mock Back-End Client
type MockClient struct {
	t    *testing.T
	fail string
}

func (c *MockClient) StoreRecord(id, record []byte) (err error) {
	if c.fail == "Store" {
		return errors.New(badBEClientMessage)
	}
	assert.Equal(c.t, idEnc, id)
	assert.Equal(c.t, recordEnc, record)
	return nil
}

func (c *MockClient) RetrieveRecord(id []byte) (record []byte, err error) {
	if c.fail == "Retrieve" {
		return nil, errors.New(badBEClientMessage)
	} else if c.fail == "RetrieveCorrupt" {
		// Corrupt nonce on encrypted record.
		return recordEnc[1:], nil
	}

	assert.Equal(c.t, idEnc, id)
	return recordEnc, nil
}

func (c *MockClient) DeleteRecord(id []byte) (err error) {
	if c.fail == "Delete" {
		return errors.New(badBEClientMessage)
	}

	assert.Equal(c.t, idEnc, id)
	return nil
}

// MakeServer() - Test Method
func TestServer_MakeServer(t *testing.T) {

	type fields struct{}
	type args struct {
		configs         map[string]string
		beClientConfigs map[string]string
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
			args: args{goodServerConfig, goodClientConfig},
			want: goodServer,
		},
		{
			name:    "should fail loading configuration",
			args:    args{badServerConfig, goodClientConfig},
			wantErr: errors.New(badServerMessage),
		},
		{
			name: "should fail building keygen",
			args: args{badKeyGenConfig, goodClientConfig},
			wantErr: &strconv.NumError{
				Func: "Atoi",
				Num:  badKeySizeStr,
				Err:  errors.New("invalid syntax"),
			},
		},
		{
			name:    "should fail generating GCM cipher",
			args:    args{badIdKeyConfig, goodClientConfig},
			wantErr: aes.KeySizeError(0),
		},
		{
			name:    "should fail building back-end client",
			args:    args{goodServerConfig, badClientConfig},
			wantErr: errors.New(badClientMessage),
		},
		{
			name:    "should fail building socket IO",
			args:    args{badPortConfig, goodClientConfig},
			wantErr: errors.New(badSocketIOMessage),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeServer(test.args.configs, test.args.beClientConfigs)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// store() - Test Method
func TestServer_store(t *testing.T) {

	type fields struct {
		keygen   utils.KeyGen
		beClient client.Client
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
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, ""},
			},
			args: args{id, record},
			want: idKey,
		},
		{
			name: "should fail generating random key",
			fields: fields{
				keygen:   &MockKeyGen{t, "RandomKey"},
				beClient: &MockClient{t, ""},
			},
			args:    args{id, record},
			wantErr: errors.New(badRandomKeyMessage),
		},
		{
			name: "should fail generating GCM cipher",
			fields: fields{
				keygen:   &MockKeyGen{t, "GetGCMCipher"},
				beClient: &MockClient{t, ""},
			},
			args:    args{id, record},
			wantErr: errors.New(badGetGCMCipherMessage),
		},
		{
			name: "should fail generating random nonce",
			fields: fields{
				keygen:   &MockKeyGen{t, "RandomNonce"},
				beClient: &MockClient{t, ""},
			},
			args:    args{id, record},
			wantErr: errors.New(badRandomNonceMessage),
		},
		{
			name: "should fail calling back-end client",
			fields: fields{
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, "Store"},
			},
			args:    args{id, record},
			wantErr: errors.New(badBEClientMessage),
		},
	}

	for _, test := range tests {
		s := &serverImpl{
			keygen:   test.fields.keygen,
			idNonce:  idNonce,
			idCipher: idCipher,
			beClient: test.fields.beClient,
		}

		t.Run(test.name, func(t *testing.T) {
			got, err := s.storeRecord(test.args.id, test.args.record)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// retrieve() - Test Methods
func TestServer_retrieve(t *testing.T) {

	type fields struct {
		keygen   utils.KeyGen
		beClient client.Client
	}
	type args struct {
		id  []byte
		key []byte
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
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, ""},
			},
			args: args{id, idKey},
			want: record,
		},
		{
			name: "should fail generating GCM cipher",
			fields: fields{
				keygen:   &MockKeyGen{t, "GetGCMCipher"},
				beClient: &MockClient{t, ""},
			},
			args:    args{id, idKey},
			wantErr: errors.New(badGetGCMCipherMessage),
		},
		{
			name: "should fail calling back-end client",
			fields: fields{
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, "Retrieve"},
			},
			args:    args{id, idKey},
			wantErr: errors.New(badBEClientMessage),
		},
		{
			name: "should fail decrypting record",
			fields: fields{
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, "RetrieveCorrupt"},
			},
			args:    args{id, idKey},
			wantErr: errors.New(badDecryptMessage),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &serverImpl{
				keygen:   test.fields.keygen,
				idNonce:  idNonce,
				idCipher: idCipher,
				beClient: test.fields.beClient,
			}

			got, err := s.retrieveRecord(test.args.id, test.args.key)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// delete() - Test Methods
func TestServer_delete(t *testing.T) {

	type fields struct {
		keygen   utils.KeyGen
		beClient client.Client
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
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, ""},
			},
			args: args{id},
		},
		{
			name: "should fail calling back-end client",
			fields: fields{
				keygen:   &MockKeyGen{t, ""},
				beClient: &MockClient{t, "Delete"},
			},
			args:    args{id},
			wantErr: errors.New(badBEClientMessage),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &serverImpl{
				keygen:   test.fields.keygen,
				idNonce:  idNonce,
				idCipher: idCipher,
				beClient: test.fields.beClient,
			}

			err := s.deleteRecord(test.args.id)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

// Respond() - Test Methods
func TestServer_Respond(t *testing.T) {

	type fields struct {
		beClient client.Client
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
				beClient: &MockClient{t, ""},
			},
			args: args{"STORE " + idHexStr + " " + recordHexStr},
			want: []byte(idKeyHexStr + "\n"),
		},
		{
			name: "should fail on StoreRecord() token count",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"STORE " + idHexStr},
			want: []byte("ERROR " + badRequest + "\n"),
		},
		{
			name: "should fail on STORE decode",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"STORE ggg ggggg"},
			want: []byte("ERROR " + badDecode + "\n"),
		},
		{
			name: "should fail on back-end client StoreRecord()",
			fields: fields{
				beClient: &MockClient{t, "Store"},
			},
			args: args{"STORE " + idHexStr + " " + recordHexStr},
			want: []byte("ERROR " + badBEClientMessage + "\n"),
		},
		{
			name: "should run RetrieveRecord() successfully",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"RETRIEVE " + idHexStr + " " + idKeyHexStr},
			want: []byte(recordHexStr + "\n"),
		},
		{
			name: "should fail on RetrieveRecord() token count",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"RETRIEVE "},
			want: []byte("ERROR " + badRequest + "\n"),
		},
		{
			name: "should fail on RETRIEVE decode",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"RETRIEVE ggg ggggg"},
			want: []byte("ERROR " + badDecode + "\n"),
		},
		{
			name: "should fail on back-end client RetrieveRecord()",
			fields: fields{
				beClient: &MockClient{t, "Retrieve"},
			},
			args: args{"RETRIEVE " + idHexStr + " " + idKeyHexStr},
			want: []byte("ERROR " + badBEClientMessage + "\n"),
		},
		{
			name: "should run DeleteRecord() successfully",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"DELETE " + idHexStr},
			want: []byte("\n"),
		},
		{
			name: "should fail on DeleteRecord() token count",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"DELETE "},
			want: []byte("ERROR " + badRequest + "\n"),
		},
		{
			name: "should fail on DELETE decode",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"DELETE ggg"},
			want: []byte("ERROR " + badDecode + "\n"),
		},
		{
			name: "should fail on back-end client DeleteRecord()",
			fields: fields{
				beClient: &MockClient{t, "Delete"},
			},
			args: args{"DELETE " + idHexStr},
			want: []byte("ERROR " + badBEClientMessage + "\n"),
		},
		{
			name: "should fail on unrecognized command",
			fields: fields{
				beClient: &MockClient{t, ""},
			},
			args: args{"FOO " + idHexStr + " " + recordHexStr},
			want: []byte("ERROR " + badRequest + "\n"),
		},
	}

	for _, test := range tests {
		s := &serverImpl{
			keygen:   &MockKeyGen{t, ""},
			idNonce:  idNonce,
			idCipher: idCipher,
			beClient: test.fields.beClient,
		}

		t.Run(test.name, func(t *testing.T) {
			got := s.Respond(test.args.message)
			assert.Equal(t, test.want, got)
		})
	}
}
