package main

import (
	"encoding/json"
	"fmt"
	"github.com/heyuhang0/DSCProject/internal/server"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type ServerConfig struct {
	Id           int
	IpInternal   string
	IpExternal   string
	PortInternal int
	PortExternal int
}

type Configuration struct {
	NumServer       int
	NumReplica      int
	NumRead         int
	NumWrite        int
	NumVirtualNodes int
	Timeout         int
	Servers         []ServerConfig
	Ids             []int
}

func main() {
	// can take: 1, 2, ..., numServer
	serverIndexStr := os.Args[1]
	serverIndex, errConvert := strconv.Atoi(serverIndexStr)
	serverIndex--
	if errConvert != nil {
		fmt.Println(errConvert)
	}
	// read configs
	configFile := "configs/default_config.json"
	jsonFile, errReadFile := os.Open(configFile)
	if errReadFile != nil {
		fmt.Println(errReadFile)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Configuration
	json.Unmarshal(byteValue, &config)
	fmt.Println(config)

	numServer := config.NumServer
	numRead := config.NumRead
	numWrite := config.NumWrite
	numReplica := config.NumReplica
	servers := config.Servers
	numVNodes := config.NumVirtualNodes

	Ids := config.Ids
	timeout := time.Millisecond * time.Duration(config.Timeout)
	log.Println(timeout)

	if serverIndex > numServer {
		os.Exit(1)
	}

	localServer := servers[serverIndex]
	nodeId := localServer.Id

	// creating server
	log.Printf("=== Server %v Starting to create server ===\n", nodeId)

	// create db
	db, errCreateDB := leveldb.OpenFile(".appdata/"+strconv.Itoa(nodeId)+"/leveldb", nil)
	if errCreateDB != nil {
		log.Fatalf("failed to initialize leveldb: %v", errCreateDB.Error())
	}
	log.Printf("Server %v Local DB created\n", nodeId)

	// listen to external and internal ports
	internalAddress := localServer.IpInternal + ":" + strconv.Itoa(localServer.PortInternal)
	lisInternal, errListenInternal := net.Listen("tcp", internalAddress)
	log.Printf("Listening internal address: %v\n", internalAddress)
	if errListenInternal != nil {
		log.Fatalf("Failed to listen to internal address %v: %v", internalAddress, errListenInternal.Error())
	}
	externalAddress := localServer.IpExternal + ":" + strconv.Itoa(localServer.PortExternal)
	lisExternal, errListenExternal := net.Listen("tcp", externalAddress)
	log.Printf("Listening external address %v", externalAddress)
	if errListenExternal != nil {
		log.Fatalf("Failed to listen to external address %v: %v", externalAddress, errListenExternal.Error())
	}

	sExternal := grpc.NewServer()
	sInternal := grpc.NewServer()
	// register to grpc
	newServer := server.NewServer(nodeId, Ids, numReplica, numRead, numWrite, numVNodes, timeout, db)

	pb.RegisterKeyValueStoreServer(sExternal, newServer)
	pb.RegisterKeyValueStoreInternalServer(sInternal, newServer)
	go func() {
		if err := sExternal.Serve(lisExternal); err != nil {
			log.Fatalf("Failed to serve external address %v: %v", externalAddress, err.Error())
		}
	}()
	go func() {
		if err := sInternal.Serve(lisInternal); err != nil {
			log.Fatalf("Failed to serve internal address %v: %v", internalAddress, err.Error())
		}
	}()

	// create connection to other service
	var clientForServer sync.Map
	var wg sync.WaitGroup
	for j := 0; j < len(servers); j++ {
		// index is the index in the list, id is the actual id
		serverClient := servers[j]
		if nodeId == serverClient.Id {
			continue
		}
		log.Printf("Server %v Creating connection to server %v", nodeId, serverClient.Id)
		wg.Add(1)
		go func(serverClient ServerConfig) {
			defer wg.Done()
			peerServerId := serverClient.Id
			address := serverClient.IpInternal + ":" + strconv.Itoa(serverClient.PortInternal)
			conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
			log.Printf("Server %v dial to server %v", nodeId, peerServerId)
			if err != nil {
				log.Fatalf("Server %v failed to dial to server %v, error: %v\n", nodeId, peerServerId, err.Error())
			}
			c := pb.NewKeyValueStoreInternalClient(conn)
			log.Printf("Connection between server %v and server %v created \n", nodeId, peerServerId)
			clientForServer.Store(peerServerId, c)
		}(serverClient)
	}
	wg.Wait()
	newServer.SetOtherServerInstance(&clientForServer)
	log.Printf("=== Finished setting server %v ===\n", nodeId)

	// sleep forever
	select {}
}
