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
	"time"
)

type Configuration struct {
	NumServer int
	NumDescendants int
	Ids    []int
	ServerPortInternal   []int
	ServerPortExternal   []int
	ServerIPInternal   []string
	ServerIPExternal   []string
}

func main() {
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

	numServer := config.NumServer
	Ids := config.Ids
	internalPorts := config.ServerPortInternal
	externalPorts := config.ServerPortExternal
	internalIP := config.ServerIPInternal
	externalIP := config.ServerIPExternal
	numDescendants := config.NumDescendants

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
	newServer := server.NewServer(nodeId, Ids, numDescendants, db)
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
	for j := 0; j < numServer; j++ {
		if serverIndex == j {
			continue
		}
		log.Println("creating connection to others ",  nodeId, Ids[j])
		address := internalIP[j] + ":" + strconv.Itoa(internalPorts[j])
		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		log.Println("dial from", nodeId, Ids[j])
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		//defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreInternalClient(conn)
		log.Println("connection between two server created ", nodeId, Ids[j])
		clientForServer = append(clientForServer, c)
	}
	newServer.SetOtherServerInstance(clientForServer)
	log.Println("finished setting server ", nodeId)


	time.Sleep(10000000000 * time.Millisecond)
}