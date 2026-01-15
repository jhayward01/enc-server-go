package server

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"
	
	"enc-server-go/pkg/v2-apis/be/service"
	"enc-server-go/pkg/utils"
)

type Server interface {
	// Start server.
	Start() (err error)
}

// Server implementation
type serverImpl struct {
	service.UnimplementedBackendServiceServer
	db utils.DB
	serverAddr string
}
	
func (s *serverImpl) Store(ctx context.Context, req *service.StoreRequest) (*service.StoreResponse, error) {
	
	if err := s.db.StoreRecord(req.Id, req.Data); err != nil {
		return nil, err
	}
	
	reply := &service.StoreResponse{
		Message: "Stored, " + req.Id,
	}
	log.Println("Server sent a store reply")
	return reply, nil
}

func (s *serverImpl) Retrieve(ctx context.Context, req *service.RetrieveRequest) (*service.RetrieveResponse, error) {
	
	record, err := s.db.RetrieveRecord(req.Id)
	if err != nil {
		return nil, err
	}
	
	reply := &service.RetrieveResponse{
		Message: "Hello, " + req.Id + "!",
		Data: record,
	}
	log.Println("Server sent a retrieve reply")
	return reply, nil
}

func (s *serverImpl) Delete(ctx context.Context, req *service.DeleteRequest) (*service.DeleteResponse, error) {
	
	if err := s.db.DeleteRecord(req.Id); err != nil {
		return nil, err
	}
	
	reply := &service.DeleteResponse{
		Message: "Deleted, " + req.Id,
	}
	log.Println("Server sent a delete reply")
	return reply, nil
}

func (s *serverImpl) Start() (err error) {
	lis, err := net.Listen("tcp", s.serverAddr)
	if err != nil {
		return errors.New("Failed to listen: " + err.Error())
	}
	
	g := grpc.NewServer()
	service.RegisterBackendServiceServer(g, s)

	log.Println("Server listening on " + s.serverAddr)
	if err := g.Serve(lis); err != nil {
		return errors.New("Failed to serve: " + err.Error())
	}
	
	return nil
}

func MakeServer(configs map[string]string) (s Server, err error) {

	// Build data store wrapper.
	db, err := utils.MakeDB(configs)
	if err != nil {
		return nil, err
	}

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs,
		[]string{"port"}); !ok {
		err = errors.New("MakeServer missing configuration " + missing)
		return nil, err
	}
	
	// Build server implementation.
	si := &serverImpl{
		db: db,
		serverAddr: "localhost:" + configs["port"],
	}
	
	return si, nil
}
