// client
package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"enc-server-go/pkg/utils"
)

type record struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Data string `json:"data"`
}

var records = []record{
	{ID: "JTH", Key: "vkAZAarLbZ6w0kmL2HJP3eU1ODCgVj4k", Data: "PAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADS"},
	{ID: "NRM", Key: "key2", Data: "PAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADS"},
	{ID: "EPB", Key: "key3", Data: "PAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADSPAYLOADS"},
}

func getMessage(c *gin.Context) {
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

type Server interface {

	// Start server.
	Start() (err error)
}

// Server implementation
type serverImpl struct {
	serverAddr string
}

func (s *serverImpl) Start() (err error) {

	router := gin.Default()
	router.GET("/records/:id", getMessage)

	router.Run(s.serverAddr)

	return err
}

func MakeServer(configs map[string]string,
	_ map[string]string) (s Server, err error) {

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs,
		[]string{"keySize", "idKeyStr", "idNonceStr", "port"}); !ok {
		err = errors.New("MakeServer missing configuration " + missing)
		return nil, err
	}

	// Build server implementation.
	si := &serverImpl{
		serverAddr: "localhost:" + configs["port"],
	}

	return si, nil
}
