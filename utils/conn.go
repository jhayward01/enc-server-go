package utils

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
)

type Conn interface {
	GetResponse(message string) (response string, err error)
}

type connImpl struct {
	serverAddr string
}

func (c *connImpl) GetResponse(message string) (response string, err error) {

	// Dial server.
	conn, err := net.Dial("tcp", c.serverAddr)
	if err != nil {
		return "", nil
	}
	defer conn.Close()

	// Write request to server.
	fmt.Fprint(conn, message)

	// Receive server response.
	response, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", nil
	}
	response = strings.TrimRight(response, "\n")

	return response, nil
}

func MakeConn(configs map[string]string) (c Conn, err error) {

	// Verify required configurations.
	if ok, missing := VerifyConfigs(configs, []string{"serverAddr"}); !ok {
		err = errors.New("MakeConn missing configuration " + missing)
		return nil, err
	}

	c = &connImpl{
		serverAddr: configs["serverAddr"],
	}
	return c, nil
}
