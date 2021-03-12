package main

import (
	"github.com/heyuhang0/DSCProject/internal/server"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	// hard coded server list
	serverIDInternal := [5]int{5000, 5002, 5003, 5004, 5005}
	serverIDExternal := [5]int{6000, 6002, 6003, 6004, 6005}

	// Set up a connection to other servers's  internal.
	for i := 0; i < len(serverIDInternal); i++ {
		// Create Level DB
		go func(i int) {
			log.Println("starting to create server", i)
			db, err := leveldb.OpenFile(".appdata/"+strconv.Itoa(serverIDInternal[i])+"/leveldb", nil)
			if err != nil {
				log.Fatalf("failed to initialize leveldb: %v", err)
			}
			log.Println("db created", i)

			// listen to external and internal ports
			internalAddress := "localhost:" + strconv.Itoa(serverIDInternal[i])
			lisInternal, err := net.Listen("tcp", internalAddress)
			log.Println("registered internal address" + internalAddress)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			externalAddress := "localhost:" + strconv.Itoa(serverIDExternal[i])
			lisExternal, err := net.Listen("tcp", externalAddress)
			log.Println("registered external address" + externalAddress)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			sExternal := grpc.NewServer()
			sInternal := grpc.NewServer()
			// register to grpc
			newServer := server.NewServer(serverIDInternal[i], serverIDInternal[:], 2, db)
			pb.RegisterKeyValueStoreServer(sExternal, newServer)
			pb.RegisterKeyValueStoreInternalServer(sInternal, newServer)
			go func() {
				if err := sExternal.Serve(lisExternal); err != nil {
					log.Fatalf("failed to serve: %v", err)
				}
			}()
			go func() {
				if err := sInternal.Serve(lisInternal); err != nil {
					log.Fatalf("failed to serve: %v", err)
				}
			}()

			// create connection to other service
			var clientForServer []pb.KeyValueStoreInternalClient
			for j := 0; j < len(serverIDInternal); j++ {
				log.Println("in the loop", i)
				if i == j {
					continue
				}
				address := "localhost:" + strconv.Itoa(serverIDInternal[j])
				conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
				log.Println("dial from", i, j)
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}
				//defer func() { _ = conn.Close() }()
				c := pb.NewKeyValueStoreInternalClient(conn)
				log.Println("connection between two server created", i, j)
				clientForServer = append(clientForServer, c)
			}
			newServer.SetOtherServerInstance(clientForServer)
			log.Println("finished setting server", i)
		}(i)
	}
	time.Sleep(100000000 * time.Millisecond)
}
