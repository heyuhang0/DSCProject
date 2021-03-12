package main

import (
	"github.com/heyuhang0/DSCProject/internal/server"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	externalPort = ":50051"
	internalPort = ":50052"
)


func main() {
	// Create Level DB
	db, err := leveldb.OpenFile(".appdata/single_node/leveldb", nil)
	if err != nil {
		log.Fatalf("failed to initialize leveldb: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create gRPC service
	lis, err := net.Listen("tcp", externalPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterKeyValueStoreServer(s, server.NewServer(db))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
