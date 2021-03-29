package main

import (
	"context"
	"flag"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/kvclient"
	"log"
)

const (
	address = "localhost:6000"
)

func main() {
	// Set up a connection to the server.
	configPath := flag.String("config", "./configs/default_client.json", "config path")
	config, err := kvclient.NewClientConfigFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	c := kvclient.NewKeyValueStoreClient(config)

	// Basic Test
	ctx := context.Background()

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
