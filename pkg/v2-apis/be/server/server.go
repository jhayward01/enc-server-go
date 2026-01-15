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

type server struct {
	service.UnimplementedBackendServiceServer
}

type Server interface {
	// Start server.
	Start() (err error)
}

// Server implementation
type serverImpl struct {
	db       utils.DB
}

func (s *server) Store(ctx context.Context, req *service.StoreRequest) (*service.StoreResponse, error) {
	reply := &service.StoreResponse{
		Message: "Hello, " + req.Id + "!",
	}
	log.Println("Server sent a store reply")
	return reply, nil
}

func (s *server) Retrieve(ctx context.Context, req *service.RetrieveRequest) (*service.RetrieveResponse, error) {
	reply := &service.RetrieveResponse{
		Message: "Hello, " + req.Id + "!",
	}
	log.Println("Server sent a retrieve reply")
	return reply, nil
}

func (s *server) Delete(ctx context.Context, req *service.DeleteRequest) (*service.DeleteResponse, error) {
	reply := &service.DeleteResponse{
		Message: "Hello, " + req.Id + "!",
	}
	log.Println("Server sent a delete reply")
	return reply, nil
}

func (s *serverImpl) Start() (err error) {
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		return errors.New("failed to listen: " + err.Error())
	}
	
	g := grpc.NewServer()
	service.RegisterBackendServiceServer(g, &server{})

	log.Println("Server listening on :8888")
	if err := g.Serve(lis); err != nil {
		return errors.New("failed to serve: " + err.Error())
	}
	
	return nil
}

func MakeServer(configs map[string]string) (s Server, err error) {

	// // Build data store wrapper.
	// db, err := utils.MakeDB(configs)
	// if err != nil {
	// 	return nil, err
	// }

	// // Build server implementation.
	// si := &serverImpl{
	// 	db: db,
	// }

	si := &serverImpl{
	}
	
	return si, nil
}