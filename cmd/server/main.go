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

type Configuration struct {
	NumServer int
	NumReplica int
	NumRead int
	NumWrite int
	Ids    []int
	ServerPortInternal   []int
	ServerPortExternal   []int
	ServerIPInternal   []string
	ServerIPExternal   []string
}

func main() {
	// can take: 1, 2, ..., numServer
	serverIndexStr := os.Args[1]
	serverIndex, err := strconv.Atoi(serverIndexStr)
	serverIndex --
	if err != nil {
		fmt.Println(err)
	}
	// read configs
	configFile := "configs/default_config.json"
	jsonFile, err := os.Open(configFile)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Configuration
	json.Unmarshal(byteValue, &config)
	fmt.Println(config)
	//os.Exit(0)

	numServer := config.NumServer
	numRead := config.NumRead
	numWrite := config.NumWrite
	numReplica := config.NumReplica
	Ids := config.Ids
	internalPorts := config.ServerPortInternal
	externalPorts := config.ServerPortExternal
	internalIP := config.ServerIPInternal
	externalIP := config.ServerIPExternal

	if serverIndex > numServer {
		os.Exit(1)
	}

	nodeId := Ids[serverIndex]

	// creating server
	log.Println("starting to create server", nodeId)

	//create db
	db, err := leveldb.OpenFile(".appdata/"+strconv.Itoa(nodeId)+"/leveldb", nil)
	if err != nil {
		log.Fatalf("failed to initialize leveldb: %v", err)
	}
	log.Println("db created", nodeId)

	// listen to external and internal ports
	internalAddress := internalIP[serverIndex] + ":" + strconv.Itoa(internalPorts[serverIndex])
	lisInternal, err := net.Listen("tcp", internalAddress)
	log.Println("registered internal address " + internalAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	externalAddress := externalIP[serverIndex] + ":" + strconv.Itoa(externalPorts[serverIndex])
	lisExternal, err := net.Listen("tcp", externalAddress)
	log.Println("registered external address " + externalAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	sExternal := grpc.NewServer()
	sInternal := grpc.NewServer()
	// register to grpc
	newServer := server.NewServer(nodeId, Ids, numReplica, numRead, numWrite, db)
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
	clientForServer := make(map[int]pb.KeyValueStoreInternalClient)
	//clientForServer := make([]pb.KeyValueStoreInternalClient, numServer - 1, numServer - 1)
	var wg sync.WaitGroup
	for j := 0; j < numServer; j++ {
		// index is the index in the list, id is the actual id
		if nodeId == Ids[j] {
			continue
		}
		log.Println("creating connection to others ",  nodeId, Ids[j])
		wg.Add(1)
		go func(otherIdIndex int){
			defer wg.Done()
			peerServerId := Ids[otherIdIndex]
			address := internalIP[otherIdIndex] + ":" + strconv.Itoa(internalPorts[otherIdIndex])
			conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
			log.Println("dial from", nodeId, peerServerId)
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			//defer func() { _ = conn.Close() }()
			c := pb.NewKeyValueStoreInternalClient(conn)
			log.Println("connection between two server created ", peerServerId)
			clientForServer[peerServerId] = c
		}(j)
	}
	wg.Wait()
	log.Println(clientForServer)
	newServer.SetOtherServerInstance(clientForServer)
	log.Println("finished setting server ", nodeId)


	time.Sleep(10000000000 * time.Millisecond)
}