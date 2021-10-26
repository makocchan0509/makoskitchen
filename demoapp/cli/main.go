package main

import (
	"context"
	"log"
	"makoskitchen/demoapp"
	"os"
	"time"

	"google.golang.org/grpc"
)

const (
	address         = "localhost:8090"
	defaultName     = "makoto"
	defaultLastName = "Mase"
)

func main() {

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := demoapp.NewGreeterClient(conn)

	name := defaultName
	lastName := defaultLastName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &demoapp.HelloRequest{Name: name, Last_Name: lastName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
