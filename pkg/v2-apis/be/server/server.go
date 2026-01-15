package main

import (
    "context"
    "log"
    "net"

    "google.golang.org/grpc"
    pb "enc-server-go/pkg/v2-apis/be/service"
)

// server is used to implement example.ExampleServiceServer
type server struct {
    pb.UnimplementedExampleServiceServer
}

// SayHello implements the SayHello RPC method
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    reply := &pb.HelloResponse{
        Message: "Hello, " + req.Name + "!",
    }
    return reply, nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    pb.RegisterExampleServiceServer(s, &server{})

    log.Println("Server listening on :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
