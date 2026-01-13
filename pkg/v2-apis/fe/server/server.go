// client
package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"enc-server-go/pkg/utils"
)

type message struct {
	ID     string `json:"id"`
	Record string `json:"record"`
}

var messages = []message{
	{ID: "1", Record: "Message1"},
	{ID: "2", Record: "Message2"},
	{ID: "3", Record: "Message3"},
}

func getMessages(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, messages)
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
	router.GET("/messages", getMessages)

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
