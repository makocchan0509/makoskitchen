package main

import (
	"context"
	"log"
	"net"

	"makoskitchen/demoapp"

	"google.golang.org/grpc"
)

type server struct {
	demoapp.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *demoapp.HelloRequest) (*demoapp.HelloReply, error) {
	log.Printf("Received: %v %v", in.Name, in.Last_Name)
	return &demoapp.HelloReply{Message: "Hello " + in.Name + in.Last_Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	demoapp.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}

}
