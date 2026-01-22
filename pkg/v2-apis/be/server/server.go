package server

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

type Server interface {
	// Start server.
	Start() (err error)
}

// Server implementation
type serverImpl struct {
	service.UnimplementedBackendServiceServer
	db         utils.DB
	serverAddr string
}

func (s *serverImpl) StoreRecord(ctx context.Context, req *service.StoreRequest) (*service.StoreResponse, error) {
	log.Println("BE server received a store request for", req.Id)

	if err := s.db.StoreRecord(req.Id, req.Data); err != nil {
		log.Println("BE server StoreRecord error:", err)
		return nil, err
	}

	return &service.StoreResponse{}, nil
}

func (s *serverImpl) RetrieveRecord(ctx context.Context, req *service.RetrieveRequest) (*service.RetrieveResponse, error) {

	log.Println("BE server received a get request for", req.Id)

	record, err := s.db.RetrieveRecord(req.Id)
	if err != nil {
		log.Println("BE server RetrieveRecord error:", err)
		return nil, err
	}

	reply := &service.RetrieveResponse{
		Data: record,
	}
	return reply, nil
}

func (s *serverImpl) DeleteRecord(ctx context.Context, req *service.DeleteRequest) (*service.DeleteResponse, error) {

	log.Println("BE server received a delete request for", req.Id)

	if err := s.db.DeleteRecord(req.Id); err != nil {
		log.Println("BE server DeleteRecord error:", err)
		return nil, err
	}

	return &service.DeleteResponse{}, nil
}

func (s *serverImpl) Start() (err error) {

	// Listen on TCP port
	lis, err := net.Listen("tcp", s.serverAddr)
	if err != nil {
		return errors.New("Failed to listen: " + err.Error())
	}

	// Create and register server
	g := grpc.NewServer()
	service.RegisterBackendServiceServer(g, s)

	log.Println("Listening and serving GRPC on", s.serverAddr)

	if err := g.Serve(lis); err != nil {
		return errors.New("Failed to serve: " + err.Error())
	}

	return nil
}

func MakeServer(configs map[string]string) (s Server, err error) {

	log.Println("BE server MakeServer with configs:", configs)

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
		db:         db,
		serverAddr: ":" + configs["port"],
	}

	return si, nil
}
