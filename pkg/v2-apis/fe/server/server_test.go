package server

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
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
		"serverAddr": ":" + port,
	}

	goodServer = &serverImpl{
		keygen:     keygen,
		idNonce:    idNonce,
		idCipher:   idCipher,
		beClient:   goodClient,
		serverAddr: ":" + port,
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

// Mock KeyGen
type mockKeyGen struct {
	getGCMCipherFn func(key []byte) (cipher.AEAD, error)
	randomKeyFn    func() ([]byte, error)
	randomNonceFn  func(nonceSize int) ([]byte, error)
}

func (m *mockKeyGen) GetGCMCipher(key []byte) (cipher.AEAD, error) {
	if m.getGCMCipherFn != nil {
		return m.getGCMCipherFn(key)
	}
	return nil, errors.New("mock error")
}

func (m *mockKeyGen) RandomKey() ([]byte, error) {
	if m.randomKeyFn != nil {
		return m.randomKeyFn()
	}
	return nil, errors.New("mock error")
}

func (m *mockKeyGen) RandomNonce(nonceSize int) ([]byte, error) {
	if m.randomNonceFn != nil {
		return m.randomNonceFn(nonceSize)
	}
	return nil, errors.New("mock error")
}

// Mock ClientBE
type mockClientBE struct {
	storeRecordFn    func(id, record []byte) error
	retrieveRecordFn func(id []byte) ([]byte, error)
	deleteRecordFn   func(id []byte) error
}

func (m *mockClientBE) StoreRecord(id, record []byte) error {
	if m.storeRecordFn != nil {
		return m.storeRecordFn(id, record)
	}
	return nil
}

func (m *mockClientBE) RetrieveRecord(id []byte) ([]byte, error) {
	if m.retrieveRecordFn != nil {
		return m.retrieveRecordFn(id)
	}
	return nil, errors.New("mock error")
}

func (m *mockClientBE) DeleteRecord(id []byte) error {
	if m.deleteRecordFn != nil {
		return m.deleteRecordFn(id)
	}
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

// postRecord() - Test Method
func TestServer_postRecord(t *testing.T) {
	tests := []struct {
		name              string
		requestBody       Record
		mockKeyGenFn      func() (utils.KeyGen)
		mockClientBEFn    func() utils.ClientBE
		expectedStatus    int
		expectedHasKey    bool
		expectedErrorMsg  string
	}{
		{
			name: "should post record successfully",
			requestBody: Record{
				ID:   idHexStr,
				Data: recordHexStr,
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					storeRecordFn: func(id, record []byte) error {
						// Verify encrypted ID and record are passed
						assert.NotNil(t, id)
						assert.NotNil(t, record)
						return nil
					},
				}
			},
			expectedStatus: http.StatusCreated,
			expectedHasKey: true,
		},
		{
			name: "should fail with invalid JSON",
			requestBody: Record{}, // Will be sent as empty/invalid
			mockKeyGenFn: func() (utils.KeyGen) {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "should fail with invalid hex ID",
			requestBody: Record{
				ID:   "invalid-hex-gg",
				Data: recordHexStr,
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{}
			},
			expectedStatus:   http.StatusBadRequest,
			expectedErrorMsg: badDecode,
		},
		{
			name: "should fail with invalid hex data",
			requestBody: Record{
				ID:   idHexStr,
				Data: "invalid-hex-gg",
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{}
			},
			expectedStatus:   http.StatusBadRequest,
			expectedErrorMsg: badDecode,
		},
		{
			name: "should fail when random key generation fails",
			requestBody: Record{
				ID:   idHexStr,
				Data: recordHexStr,
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return &mockKeyGen{
					getGCMCipherFn: func(key []byte) (cipher.AEAD, error) {
						return idCipher, nil
					},
					randomKeyFn: func() ([]byte, error) {
						return nil, errors.New("random key generation failed")
					},
				}
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "should fail when GCM cipher creation fails",
			requestBody: Record{
				ID:   idHexStr,
				Data: recordHexStr,
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return &mockKeyGen{
					getGCMCipherFn: func(key []byte) (cipher.AEAD, error) {
						return nil, errors.New("invalid key size")
					},
					randomKeyFn: func() ([]byte, error) {
						return keygen.RandomKey()
					},
				}
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "should fail when nonce generation fails",
			requestBody: Record{
				ID:   idHexStr,
				Data: recordHexStr,
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return &mockKeyGen{
					getGCMCipherFn: func(key []byte) (cipher.AEAD, error) {
						return idCipher, nil
					},
					randomKeyFn: func() ([]byte, error) {
						return keygen.RandomKey()
					},
					randomNonceFn: func(nonceSize int) ([]byte, error) {
						return nil, errors.New("nonce generation failed")
					},
				}
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "should fail when backend client fails to store record",
			requestBody: Record{
				ID:   idHexStr,
				Data: recordHexStr,
			},
			mockKeyGenFn: func() (utils.KeyGen) {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					storeRecordFn: func(id, record []byte) error {
						return errors.New("backend storage failed")
					},
				}
			},
			expectedStatus: http.StatusBadGateway,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kg := test.mockKeyGenFn()
			mockKg, _ := kg.(*mockKeyGen)
			if mockKg == nil {
				mockKg = &mockKeyGen{
					getGCMCipherFn: func(key []byte) (cipher.AEAD, error) {
						return keygen.GetGCMCipher(key)
					},
				}
			}

			idCipherTest, _ := kg.GetGCMCipher([]byte(idKeyStr))
			server := &serverImpl{
				keygen:     kg,
				idNonce:    idNonce,
				idCipher:   idCipherTest,
				beClient:   test.mockClientBEFn(),
				serverAddr: ":" + port,
			}

			// Create request
			body, _ := json.Marshal(test.requestBody)
			req, _ := http.NewRequest("POST", "/records", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Call handler
			server.postRecord(ctx)

			// Verify status code
			assert.Equal(t, test.expectedStatus, w.Code)

			// If success, verify response has key
			if test.expectedStatus == http.StatusCreated && test.expectedHasKey {
				var resp Record
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NotEmpty(t, resp.Key)
				assert.Equal(t, test.requestBody.ID, resp.ID)
			}

			// If error, verify error message
			if test.expectedStatus >= 400 && test.expectedErrorMsg != "" {
				assert.Contains(t, w.Body.String(), test.expectedErrorMsg)
			}
		})
	}
}

// getRecord() - Test Method
func TestServer_getRecord(t *testing.T) {
	tests := []struct {
		name             string
		idParam          string
		keyParam         string
		mockKeyGenFn     func() utils.KeyGen
		mockClientBEFn   func() utils.ClientBE
		expectedStatus   int
		expectedErrorMsg string
		validateData     func(data []byte, t *testing.T)
	}{
		{
			name:    "should get record successfully",
			idParam: idHexStr,
			keyParam: func() string {
				key, _ := keygen.RandomKey()
				return hex.EncodeToString(key)
			}(),
			mockKeyGenFn: func() utils.KeyGen {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					retrieveRecordFn: func(id []byte) ([]byte, error) {
						// Verify encrypted ID is passed
						assert.NotNil(t, id)
						// Return the encrypted record
						key, _ := keygen.RandomKey()
						cipher, _ := keygen.GetGCMCipher(key)
						nonce, _ := keygen.RandomNonce(cipher.NonceSize())
						return cipher.Seal(nonce, nonce, record, nil), nil
					},
				}
			},
			expectedStatus: http.StatusOK,
			validateData: func(data []byte, t *testing.T) {
				// Note: decrypted data will be the raw bytes, comparison depends on implementation
				assert.NotNil(t, data)
			},
		},
		{
			name:           "should fail when key query parameter missing",
			idParam:        idHexStr,
			keyParam:       "",
			mockKeyGenFn:   func() utils.KeyGen { return keygen },
			mockClientBEFn: func() utils.ClientBE { return &mockClientBE{} },
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "should fail with invalid hex ID",
			idParam:        "invalid-hex-gg",
			keyParam:       hex.EncodeToString(make([]byte, 32)),
			mockKeyGenFn:   func() utils.KeyGen { return keygen },
			mockClientBEFn: func() utils.ClientBE { return &mockClientBE{} },
			expectedStatus: http.StatusBadRequest,
			expectedErrorMsg: badDecode,
		},
		{
			name:           "should fail with invalid hex key",
			idParam:        idHexStr,
			keyParam:       "invalid-hex-gg",
			mockKeyGenFn:   func() utils.KeyGen { return keygen },
			mockClientBEFn: func() utils.ClientBE { return &mockClientBE{} },
			expectedStatus: http.StatusBadRequest,
			expectedErrorMsg: badDecode,
		},
		{
			name:    "should fail when GCM cipher creation fails",
			idParam: idHexStr,
			keyParam: hex.EncodeToString(make([]byte, 32)),
			mockKeyGenFn: func() utils.KeyGen {
				return &mockKeyGen{
					getGCMCipherFn: func(key []byte) (cipher.AEAD, error) {
						return nil, errors.New("invalid key size")
					},
				}
			},
			mockClientBEFn: func() utils.ClientBE { return &mockClientBE{} },
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "should fail when backend client fails to retrieve record",
			idParam: idHexStr,
			keyParam: hex.EncodeToString(make([]byte, 32)),
			mockKeyGenFn: func() utils.KeyGen {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					retrieveRecordFn: func(id []byte) ([]byte, error) {
						return nil, errors.New("backend retrieval failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:    "should fail when decryption fails (corrupted data)",
			idParam: idHexStr,
			keyParam: hex.EncodeToString(make([]byte, 32)),
			mockKeyGenFn: func() utils.KeyGen {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					retrieveRecordFn: func(id []byte) ([]byte, error) {
						// Return corrupted data that won't decrypt
						return []byte("corrupted-data-that-wont-decrypt"), nil
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kg := test.mockKeyGenFn()
			idCipherTest, _ := kg.GetGCMCipher([]byte(idKeyStr))

			server := &serverImpl{
				keygen:     kg,
				idNonce:    idNonce,
				idCipher:   idCipherTest,
				beClient:   test.mockClientBEFn(),
				serverAddr: ":" + port,
			}

			// Create request with path and query parameters
			url := "/records/" + test.idParam
			if test.keyParam != "" {
				url += "?key=" + test.keyParam
			}
			req, _ := http.NewRequest("GET", url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = gin.Params{
				{Key: "id", Value: test.idParam},
			}

			// Call handler
			server.getRecord(ctx)

			// Verify status code
			assert.Equal(t, test.expectedStatus, w.Code)

			// If success, verify response structure
			if test.expectedStatus == http.StatusOK {
				var resp Record
				_ = json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, test.idParam, resp.ID)
				assert.Equal(t, test.keyParam, resp.Key)
				if test.validateData != nil {
					test.validateData([]byte(resp.Data), t)
				}
			}

			// If error, verify error message
			if test.expectedStatus >= 400 && test.expectedErrorMsg != "" {
				assert.Contains(t, w.Body.String(), test.expectedErrorMsg)
			}
		})
	}
}

// deleteRecord() - Test Method
func TestServer_deleteRecord(t *testing.T) {
	tests := []struct {
		name             string
		idParam          string
		mockKeyGenFn     func() utils.KeyGen
		mockClientBEFn   func() utils.ClientBE
		expectedStatus   int
		expectedErrorMsg string
	}{
		{
			name:    "should delete record successfully",
			idParam: idHexStr,
			mockKeyGenFn: func() utils.KeyGen {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					deleteRecordFn: func(id []byte) error {
						// Verify encrypted ID is passed
						assert.NotNil(t, id)
						return nil
					},
				}
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "should fail with invalid hex ID",
			idParam:        "invalid-hex-gg",
			mockKeyGenFn:   func() utils.KeyGen { return keygen },
			mockClientBEFn: func() utils.ClientBE { return &mockClientBE{} },
			expectedStatus: http.StatusBadRequest,
			expectedErrorMsg: badDecode,
		},
		{
			name:    "should fail when backend client fails to delete record",
			idParam: idHexStr,
			mockKeyGenFn: func() utils.KeyGen {
				return keygen
			},
			mockClientBEFn: func() utils.ClientBE {
				return &mockClientBE{
					deleteRecordFn: func(id []byte) error {
						return errors.New("backend deletion failed")
					},
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kg := test.mockKeyGenFn()
			idCipherTest, _ := kg.GetGCMCipher([]byte(idKeyStr))

			server := &serverImpl{
				keygen:     kg,
				idNonce:    idNonce,
				idCipher:   idCipherTest,
				beClient:   test.mockClientBEFn(),
				serverAddr: ":" + port,
			}

			// Create request
			req, _ := http.NewRequest("DELETE", "/records/"+test.idParam, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Params = gin.Params{
				{Key: "id", Value: test.idParam},
			}

			// Call handler
			server.deleteRecord(ctx)

			// Verify status code
			assert.Equal(t, test.expectedStatus, w.Code)

			// If error, verify error message
			if test.expectedStatus >= 400 && test.expectedErrorMsg != "" {
				assert.Contains(t, w.Body.String(), test.expectedErrorMsg)
			}
		})
	}
}
