// client
package server

import (
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/client"
)

type record struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Data string `json:"data"`
}

var records = []record{}

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
	var newRecord record
	if err := c.BindJSON(&newRecord); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	id, data := []byte(newRecord.ID), []byte(newRecord.Data)

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

	newRecord.Key = hex.EncodeToString(key)
	c.IndentedJSON(http.StatusCreated, newRecord)
}

func (s *serverImpl) getRecord(c *gin.Context) {
	id := c.Param("id")
	keyParam := c.Query("key")

	if keyParam == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "key not defined"})
		return
	}

	for _, a := range records {
		if a.ID == id {
			if keyParam != a.Key {
				c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "incorrect key"})
				return
			}
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "record not found"})
}

func (s *serverImpl) deleteRecord(c *gin.Context) {
	id := c.Param("id")

	for i, a := range records {
		if a.ID == id {
			records = append(records[:i], records[i+1:]...)
			c.IndentedJSON(http.StatusAccepted, gin.H{"message": "record not found"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "record not found"})
}

func (s *serverImpl) Start() (err error) {

	router := gin.Default()

	router.POST("/records", s.postRecord)

	router.GET("/records/:id", s.getRecord)

	router.DELETE("/records/:id", s.deleteRecord)

	router.Run(s.serverAddr)

	return err
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
