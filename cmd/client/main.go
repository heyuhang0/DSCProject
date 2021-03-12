package main

import (
	"context"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"google.golang.org/grpc"
	"log"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() { _ = conn.Close() }()
	c := pb.NewKeyValueStoreClient(conn)

	// Basic Test
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.Put(ctx, &pb.PutRequest{
		Key:    []byte("hello"),
		Object: []byte("world"),
	})
	if err != nil {
		log.Fatalf("failed to put: %v", err)
	}
	resp, err := c.Get(ctx, &pb.GetRequest{Key: []byte("hello")})
	if err != nil {
		log.Fatalf("failed to put: %v", err)
	}
	log.Printf("hello %v!", string(resp.Object))
}
