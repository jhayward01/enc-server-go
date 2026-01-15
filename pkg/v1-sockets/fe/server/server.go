// client
package server

import (
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"log"
	"strings"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v1-sockets/be/client"
)

// Server implementation
type serverImpl struct {
	keygen utils.KeyGen

	idNonce  []byte
	idCipher cipher.AEAD

	beClient utils.ClientBE

	socketIO *utils.SocketIO
}

func decodeHexArray(arr []string) (result [][]byte, err error) {

	// Decode hex strings to byte arrays.
	result = make([][]byte, len(arr))
	for i, s := range arr {
		var h []byte
		if h, err = hex.DecodeString(s); err != nil {
			return nil, err
		}
		result[i] = h
	}

	return result, nil
}

func (s *serverImpl) storeRecord(id, record []byte) (key []byte, err error) {

	// Generate cipher entry for ID.
	idEncrypt := s.idCipher.Seal(s.idNonce, s.idNonce, id, nil)

	// Generate random AES key.
	if key, err = s.keygen.RandomKey(); err != nil {
		return nil, err
	}

	// Generate key, cipher, and nonce for record.
	cipher, err := s.keygen.GetGCMCipher(key)
	if err != nil {
		return nil, err
	}

	// Randomly generate nonce (initialization vector).
	nonce, err := s.keygen.RandomNonce(cipher.NonceSize())
	if err != nil {
		return nil, err
	}

	// Generate cipher entry for record. Place in data store.
	recordEncrypt := cipher.Seal(nonce, nonce, record, nil)
	if err = s.beClient.StoreRecord(idEncrypt, recordEncrypt); err != nil {
		return nil, err
	}

	return key, nil
}

func (s *serverImpl) retrieveRecord(id, key []byte) (record []byte, err error) {

	// Generate fixed cipher entry for ID for lookup.
	idEncrypt := s.idCipher.Seal(s.idNonce, s.idNonce, id, nil)

	// Generate cipher for record.
	cipher, err := s.keygen.GetGCMCipher(key)
	if err != nil {
		return nil, err
	}

	// Retrieve record from data store.
	recordEncrypt, err := s.beClient.RetrieveRecord(idEncrypt)
	if err != nil {
		return nil, err
	}

	// Decrypt record from cipher entry.
	nonce := recordEncrypt[:cipher.NonceSize()]
	remainder := recordEncrypt[cipher.NonceSize():]
	if record, err = cipher.Open(nil, nonce, remainder, nil); err != nil {
		return nil, err
	}

	return record, err
}

func (s *serverImpl) deleteRecord(id []byte) (err error) {

	// Generate fixed cipher entry for ID for lookup.
	idEncrypt := s.idCipher.Seal(s.idNonce, s.idNonce, id, nil)

	// Delete record from data store.
	err = s.beClient.DeleteRecord(idEncrypt)
	if err != nil {
		return err
	}

	return nil
}

func (s *serverImpl) Respond(message string) (response []byte) {

	message = strings.TrimRight(message, " \n")
	log.Println(message)

	// Split message.
	fields := strings.Split(message, " ")

	// Compose response.
	switch fields[0] {
	case "STORE":
		const expectedFields = 3
		if len(fields) != expectedFields {
			response = []byte("ERROR Malformed request\n")
			return response
		}

		decodedBytes, err := decodeHexArray(fields[1:])
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		id, record := decodedBytes[0], decodedBytes[1]

		key, err := s.storeRecord(id, record)
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		response = []byte(hex.EncodeToString(key) + "\n")

	case "RETRIEVE":
		const expectedFields = 3
		if len(fields) != expectedFields {
			response = []byte("ERROR Malformed request\n")
			return response
		}

		decodedBytes, err := decodeHexArray(fields[1:])
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		id, key := decodedBytes[0], decodedBytes[1]

		record, err := s.retrieveRecord(id, key)
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		response = []byte(hex.EncodeToString(record) + "\n")

	case "DELETE":
		const expectedFields = 2
		if len(fields) != expectedFields {
			response = []byte("ERROR Malformed request\n")
			return response
		}

		decodedBytes, err := decodeHexArray(fields[1:])
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		id := decodedBytes[0]

		err = s.deleteRecord(id)
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		response = []byte("\n")

	default:
		response = []byte("ERROR Malformed request\n")
	}

	return response
}

func (s *serverImpl) Start() (err error) {

	// Start socket IO.
	if err = s.socketIO.Start(); err != nil {
		return err
	}
	return nil
}

func MakeServer(configs map[string]string,
	beClientConfigs map[string]string) (s utils.Server, err error) {

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs,
		[]string{"keySize", "idKeyStr", "idNonceStr", "port"}); !ok {
		err = errors.New("MakeServer missing configuration " + missing)
		return nil, err
	}

	// Initialize fields that require error handling.
	keygen, err := utils.MakeKeyGen(configs)
	if err != nil {
		return nil, err
	}

	idCipher, err := keygen.GetGCMCipher([]byte(configs["idKeyStr"]))
	if err != nil {
		return nil, err
	}

	beClient, err := client.MakeClient(beClientConfigs)
	if err != nil {
		return nil, err
	}

	// Build from server implementation.
	si := &serverImpl{
		keygen: keygen,

		idNonce:  []byte(configs["idNonceStr"]),
		idCipher: idCipher,

		beClient: beClient,
	}

	// Create socket IO with reference to response function.
	if si.socketIO, err = utils.MakeSocketIO(configs, si); err != nil {
		return nil, err
	}

	return si, nil
}
