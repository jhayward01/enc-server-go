package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    pb "enc-server-go/pkg/v2-apis/be/service" 
)

func main() {
    // Connect to the server
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()

    c := pb.NewExampleServiceClient(conn)

    // Call SayHello
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    req := &pb.HelloRequest{Name: "World"}
    res, err := c.SayHello(ctx, req)
    if err != nil {
        log.Fatalf("could not greet: %v", err)
    }

    log.Printf("Greeting: %s", res.Message)
}
