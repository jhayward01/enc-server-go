// client
package server

import (
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/client"
)

type Record struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Data string `json:"data"`
}

type Server interface {

	// Start server.
	Start() (err error)
}

// Server implementation
type serverImpl struct {
	keygen utils.KeyGen

	idNonce  []byte
	idCipher cipher.AEAD

	beClient utils.ClientBE

	serverAddr string
}

func (s *serverImpl) postRecord(c *gin.Context) {

	// Extract record ID and data
	var newRecord Record
	if err := c.BindJSON(&newRecord); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	log.Println("FE server received a post request for", newRecord.ID)

	// Extract ID to hex
	id, err := hex.DecodeString(newRecord.ID)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Extract data to hex
	data, err := hex.DecodeString(newRecord.Data)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Generate cipher entry for ID.
	idEncrypt := s.idCipher.Seal(s.idNonce, s.idNonce, id, nil)

	// Generate random AES key.
	key, err := s.keygen.RandomKey()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Generate key, cipher, and nonce for record.
	cipher, err := s.keygen.GetGCMCipher(key)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Randomly generate nonce (initialization vector).
	nonce, err := s.keygen.RandomNonce(cipher.NonceSize())
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Generate cipher entry for record. Place in data store.
	recordEncrypt := cipher.Seal(nonce, nonce, data, nil)
	if err := s.beClient.StoreRecord(idEncrypt, recordEncrypt); err != nil {
		c.IndentedJSON(http.StatusBadGateway, gin.H{"message": err.Error()})
		return
	}

	// Return new record with key
	newRecord.Key = hex.EncodeToString(key)
	c.IndentedJSON(http.StatusCreated, newRecord)
}

func (s *serverImpl) getRecord(c *gin.Context) {
	idStr := c.Param("id")
	keyStr := c.Query("key")

	log.Println("FE server received a get request for", idStr)

	// Verify paramaters
	if keyStr == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "key not defined"})
		return
	}

	// Extract ID to hex
	id, err := hex.DecodeString(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Extract key to hex
	key, err := hex.DecodeString(keyStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Generate fixed cipher entry for ID for lookup.
	idEncrypt := s.idCipher.Seal(s.idNonce, s.idNonce, id, nil)

	// Generate cipher for record.
	cipher, err := s.keygen.GetGCMCipher(key)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Retrieve record from data store.
	recordEncrypt, err := s.beClient.RetrieveRecord(idEncrypt)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Decrypt record from cipher entry.
	var data []byte
	nonce := recordEncrypt[:cipher.NonceSize()]
	remainder := recordEncrypt[cipher.NonceSize():]
	if data, err = cipher.Open(nil, nonce, remainder, nil); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Return retrieved record with key
	retrievedRecord := Record{
		ID:   idStr,
		Data: string(data),
		Key:  keyStr,
	}
	c.IndentedJSON(http.StatusOK, retrievedRecord)
}

func (s *serverImpl) deleteRecord(c *gin.Context) {
	idStr := c.Param("id")

	log.Println("FE server received a delete request for", idStr)

	// Extract ID to hex
	id, err := hex.DecodeString(idStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Generate fixed cipher entry for ID for lookup.
	idEncrypt := s.idCipher.Seal(s.idNonce, s.idNonce, id, nil)

	// Delete record from data store.
	if err = s.beClient.DeleteRecord(idEncrypt); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Return accepted status
	c.IndentedJSON(http.StatusAccepted, nil)
}

func (s *serverImpl) Start() (err error) {

	// Create Gin router
	router := gin.Default()

	// RESTful endpoints
	router.POST("/records", s.postRecord)
	router.GET("/records/:id", s.getRecord)
	router.DELETE("/records/:id", s.deleteRecord)

	// Start router
	router.Run(s.serverAddr)

	return nil
}

func MakeServer(configs map[string]string,
	beClientConfigs map[string]string) (s Server, err error) {

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

	// Build server implementation.
	si := &serverImpl{
		keygen: keygen,

		idNonce:  []byte(configs["idNonceStr"]),
		idCipher: idCipher,

		beClient: beClient,

		serverAddr: "localhost:" + configs["port"],
	}

	return si, nil
}
