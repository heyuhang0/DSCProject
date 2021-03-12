package main

import (
	"github.com/heyuhang0/DSCProject/internal/server"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

const (
	externalPort = ":50051"
	internalPort = ":50052"
)

type ServerCfg struct {
	clients [][]pb.KeyValueStoreInternalClient
	dbs []*leveldb.DB
}

func main() {
	// hard coded server list
	serverIDInternal := [5]int{5000, 5002, 5003, 5004, 5005}
	serverIDExternal := [5]int{5000, 5002, 5003, 5004, 5005}

	// Set up a connection to other servers's  internal.
	for i := 0; i < len(serverIDInternal); i++{
		// Create Level DB
		db, err := leveldb.OpenFile(".appdata/" + strconv.Itoa(serverIDInternal[i]) + "/leveldb", nil)
		if err != nil {
			log.Fatalf("failed to initialize leveldb: %v", err)
		}
		defer func() { _ = db.Close() }()

		// create connection to other service
		var clientForServer []pb.KeyValueStoreInternalClient
		for j := 0; j < len(serverIDInternal); j ++ {
			if i == j{
				continue
			}
			address := "localhost:" + strconv.Itoa(serverIDInternal[j])
			conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer func() { _ = conn.Close() }()
			c := pb.NewKeyValueStoreInternalClient(conn)
			clientForServer = append(clientForServer, c)
		}

		// listen to external and internal ports
		internalAddress := "localhost:" + strconv.Itoa(serverIDInternal[i])
		lis, err := net.Listen("tcp", internalAddress)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		externalAddress := "localhost:" + strconv.Itoa(serverIDExternal[i])
		lis, err = net.Listen("tcp", externalAddress)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		// register to grpc
		newServer := server.NewServer(serverIDInternal[i], serverIDInternal[:], clientForServer, 3, db)
		pb.RegisterKeyValueStoreServer(s, newServer)
		pb.RegisterKeyValueStoreInternalServer(s, newServer)

	}

	// Create gRPC service
	for i := 0; i < len(serverIDInternal); i ++ {

	}

}
