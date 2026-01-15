package client

// import (
// 	"context"
// 	"log"
// 	"time"

// 	pb "enc-server-go/pkg/v2-apis/be/service"
// 	"google.golang.org/grpc"
// )

// func main() {
// 	// Connect to the server
// 	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()

// 	c := pb.NewExampleServiceClient(conn)

// 	// Call SayHello
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	req := &pb.HelloRequest{Name: "World"}
// 	res, err := c.SayHello(ctx, req)
// 	if err != nil {
// 		log.Fatalf("could not greet: %v", err)
// 	}

// 	log.Printf("Greeting: %s", res.Message)
// }

import (
	"context"
	"errors"
	"time"
	
	"google.golang.org/grpc"

	"enc-server-go/pkg/utils"
	"enc-server-go/pkg/v2-apis/be/service"
)

// Client implementation.
type clientImpl struct {
	serverAddr string
}

func (c *clientImpl) StoreRecord(id, record []byte) (err error) {
	
	conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return errors.New("Error connecting to backend server: " + err.Error())
	}
	defer conn.Close()
	
	s := service.NewExampleServiceClient(conn)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &service.HelloRequest{Name: "World"}
	_, err = s.SayHello(ctx, req)
	if err != nil {
		return errors.New("Could not send message: " + err.Error())
	}

	return nil
}

func (c *clientImpl) RetrieveRecord(id []byte) (record []byte, err error) {
	return nil, nil
}


func (c *clientImpl) DeleteRecord(id []byte) (err error) {
	return nil
}

func MakeClient(configs map[string]string) (c utils.ClientBE, err error) {

	// Verify required configurations.
	if ok, missing := utils.VerifyConfigs(configs, []string{"serverAddr"}); !ok {
		err = errors.New("MakeClient missing configuration " + missing)
		return nil, err
	}

	// Build client implementation.
	c = &clientImpl{
		serverAddr: configs["serverAddr"],
	}

	return c, nil
}
