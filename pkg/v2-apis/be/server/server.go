package main

import (
	"context"
	"log"
	"net"

	pb "enc-server-go/pkg/v2-apis/be/service"
	"google.golang.org/grpc"
)

// server is used to implement example.ExampleServiceServer
type server struct {
	pb.UnimplementedExampleServiceServer
}

func (s *server) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResponse, error) {
	reply := &pb.StoreResponse{
		Message: "Hello, " + req.Id + "!",
	}
	log.Println("Server sent reply")
	return reply, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterExampleServiceServer(s, &server{})

	log.Println("Server listening on :8888")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
