package server

import (
	"crypto/aes"
	"encoding/hex"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/client"
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
		"serverAddr": "localhost:" + port,
	}

	goodServer = &serverImpl{
		keygen:     keygen,
		idNonce:    idNonce,
		idCipher:   idCipher,
		beClient:   goodClient,
		serverAddr: "localhost:" + port,
	}

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
const badRandomKeyMessage = "KeyGen.RandomKey error"
const badGetGCMCipherMessage = "KeyGen.GetGCMKey error"
const badRandomNonceMessage = "KeyGen.RandomNonce error"
const badBEClientMessage = "Back-end client error"
const badDecryptMessage = "cipher: message authentication failed"
const badRequest = "Malformed request"
const badDecode = "encoding/hex: invalid byte: U+0067 'g'"

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
		want    utils.Server
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MakeServer(test.args.configs, test.args.beClientConfigs)
			assert.Equal(t, test.want, got)
			assert.Equal(t, test.wantErr, err)
		})
	}
}
