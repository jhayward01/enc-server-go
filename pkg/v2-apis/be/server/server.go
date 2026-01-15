package main

import (
	"context"
	"log"
	"net"

	pb "enc-server-go/pkg/v2-apis/be/service"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedBackendServiceServer
}

func (s *server) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResponse, error) {
	reply := &pb.StoreResponse{
		Message: "Hello, " + req.Id + "!",
	}
	log.Println("Server sent a store reply")
	return reply, nil
}

func (s *server) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	reply := &pb.RetrieveResponse{
		Message: "Hello, " + req.Id + "!",
	}
	log.Println("Server sent a retrieve reply")
	return reply, nil
}

func (s *server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	reply := &pb.DeleteResponse{
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
	pb.RegisterBackendServiceServer(s, &server{})

	log.Println("Server listening on :8888")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
