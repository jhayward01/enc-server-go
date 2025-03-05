package utils

import (
	"bufio"
	"errors"
	"log"
	"net"
)

// Responding objects must fulfill this interface.
type Responder interface {
	Respond(message string) (response []byte)
}

// SocketIO configuration.
type SocketIO struct {
	port      string
	responder Responder
}

func (s *SocketIO) session(l net.Listener) (err error) {

	// Accept client connection
	c, err := l.Accept()
	if err != nil {
		return err
	}
	defer c.Close()
	log.Println("Client session opened")

	for {
		// Read transmitted messages.
		var message string
		if message, err = bufio.NewReader(c).ReadString('\n'); err != nil {
			if err.Error() == "EOF" {
				log.Println("Client session closed")
				return nil
			}
		}

		// Compose and transmit response.
		response := s.responder.Respond(message)
		if _, err = c.Write(response); err != nil {
			return err
		}
	}
}

func (s *SocketIO) Start() (err error) {

	// Start listener on port.
	log.Println("Starting server.")
	l, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	defer l.Close()

	// Process client sessions.
	for {
		if err = s.session(l); err != nil {
			return err
		}
	}
}

func MakeSocketIO(configs map[string]string, responder Responder) (s *SocketIO, err error) {

	// Verify required configurations.
	if ok, missing := VerifyConfigs(configs, []string{"port"}); !ok {
		err = errors.New("MakeSocketIO missing configuration " + missing)
		return nil, err
	}

	// Verify required configurations.
	if configs["port"] == "" {
		err = errors.New("MakeSocketIO cannot be configured with empty port")
		return nil, err
	}

	// Build socket IO with responder reference.
	s = &SocketIO{
		port:      configs["port"],
		responder: responder,
	}

	return s, nil
}
