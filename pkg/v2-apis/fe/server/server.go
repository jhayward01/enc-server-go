// client
package server

import (
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
	db utils.DB
}

func (s *serverImpl) Start() (err error) {

	router := gin.Default()
	router.GET("/messages", getMessages)

	router.Run("localhost:8080")

	return err
}

func MakeServer(configs map[string]string,
	_ map[string]string) (s Server, err error) {

	// Build data store wrapper.
	db, err := utils.MakeDB(configs)
	if err != nil {
		return nil, err
	}

	// Build server implementation.
	si := &serverImpl{
		db: db,
	}

	return si, nil
}
