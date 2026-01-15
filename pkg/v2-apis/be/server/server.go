package main

import (
	"context"
	"log"
	"net"

	"enc-server-go/pkg/v2-apis/be/service"
	"google.golang.org/grpc"
)

type server struct {
	service.UnimplementedBackendServiceServer
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

func main() {
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	service.RegisterBackendServiceServer(s, &server{})

	log.Println("Server listening on :8888")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
