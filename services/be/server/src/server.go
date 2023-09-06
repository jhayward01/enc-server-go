// client
package server

import (
	"log"
	"strings"

	"enc-server-go/utils"
)

type Server interface {

	// Start server.
	Start() (err error)
}

// Server implementation
type serverImpl struct {
	db       utils.DB
	socketIO *utils.SocketIO
}

func (s *serverImpl) store(id, record string) (err error) {

	// Call data store wrapper store method.
	if err = s.db.StoreRecord(id, record); err != nil {
		return err
	}

	return nil
}

func (s *serverImpl) retrieve(id string) (record string, err error) {

	// Call data store wrapper retrieve method.
	if record, err = s.db.RetrieveRecord(id); err != nil {
		return "", err
	}

	return record, nil
}

func (s *serverImpl) delete(id string) (err error) {

	// Call data store wrapper delete method.
	if err = s.db.DeleteRecord(id); err != nil {
		return err
	}

	return nil
}

func (s *serverImpl) Respond(message string) (response []byte) {

	log.Println(message)

	// Split message.
	fields := strings.Split(strings.TrimRight(message, " \n"), " ")

	// Compose response.
	switch fields[0] {
	case "STORE":
		const expectedFields = 3
		if len(fields) != expectedFields {
			response = []byte("ERROR Malformed request\n")
			return response
		}
		id, record := fields[1], fields[2]

		if err := s.store(id, record); err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		response = []byte("SUCCESS\n")

	case "RETRIEVE":
		const expectedFields = 2
		if len(fields) != expectedFields {
			response = []byte("ERROR Malformed request\n")
			return response
		}
		id := fields[1]

		record, err := s.retrieve(id)
		if err != nil {
			response = []byte("ERROR " + err.Error() + "\n")
			return response
		}
		response = []byte(record + "\n")

	case "DELETE":
		const expectedFields = 2
		if len(fields) != expectedFields {
			response = []byte("ERROR Malformed request\n")
			return response
		}
		id := fields[1]

		err := s.delete(id)
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
	err = s.socketIO.Start()
	return err
}

func MakeServer(configs map[string]string) (s Server, err error) {

	// Build data store wrapper.
	db, err := utils.MakeDB(configs)
	if err != nil {
		return nil, err
	}

	// Build server implementation.
	si := &serverImpl{
		db: db,
	}

	// Create socket IO with reference to response function.
	if si.socketIO, err = utils.MakeSocketIO(configs, si); err != nil {
		return nil, err
	}

	return si, nil
}
