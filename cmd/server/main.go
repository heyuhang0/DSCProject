package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/heyuhang0/DSCProject/internal/nodemgr"
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

type ServerConfig struct {
	Id           uint64
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
	Servers         []*ServerConfig
}

func main() {
	// parse arguments
	serverIdx := flag.Int("index", 1, "server index")
	serverConfig := flag.String("config", "./configs/default_config.json", "config file path")
	flag.Parse()
	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(1)
	}

	// read configs
	jsonFile, err := os.Open(*serverConfig)
	if err != nil {
		log.Fatal(err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	_ = jsonFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	var config Configuration
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Fatal(err)
	}

	numServer := config.NumServer
	numRead := config.NumRead
	numWrite := config.NumWrite
	numReplica := config.NumReplica
	servers := config.Servers
	numVNodes := config.NumVirtualNodes
	timeout := time.Millisecond * time.Duration(config.Timeout)

	if *serverIdx <= 0 || *serverIdx > numServer {
		log.Fatalf("Server index %v out of range [1, %v]", *serverIdx, numServer)
	}

	localServer := servers[*serverIdx-1]
	nodeId := localServer.Id

	// creating server
	log.Printf("=== Server %v Starting to create server ===\n", nodeId)

	// create db
	db, err := leveldb.OpenFile(fmt.Sprintf(".appdata/%d/leveldb", nodeId), nil)
	if err != nil {
		log.Fatalf("failed to initialize leveldb: %v", err)
	}
	log.Printf("Server %v Local DB created\n", nodeId)

	// create node manager
	nodeManager := nodemgr.NewManager(numVNodes)
	for _, nodeConfig := range servers {
		nodeManager.UpdateNode(&nodemgr.NodeInfo{
			ID:              nodeConfig.Id,
			Alive:           true,
			InternalAddress: fmt.Sprintf("%v:%v", nodeConfig.IpInternal, nodeConfig.PortInternal),
			ExternalAddress: fmt.Sprintf("%v:%v", nodeConfig.IpExternal, nodeConfig.PortExternal),
			Version:         time.Now().UnixNano(),
		})
	}

	// create server instance
	newServer := server.NewServer(nodeId, numReplica, numRead, numVNodes, numWrite, timeout, nodeManager, db)

	// listen to external and internal ports
	internalAddress := localServer.IpInternal + ":" + strconv.Itoa(localServer.PortInternal)
	lisInternal, err := net.Listen("tcp", internalAddress)
	log.Printf("Listening internal address: %v\n", internalAddress)
	if err != nil {
		log.Fatalf("Failed to listen to internal address %v: %v", internalAddress, err)
	}

	externalAddress := localServer.IpExternal + ":" + strconv.Itoa(localServer.PortExternal)
	lisExternal, err := net.Listen("tcp", externalAddress)
	log.Printf("Listening external address: %v", externalAddress)
	if err != nil {
		log.Fatalf("Failed to listen to external address %v: %v", externalAddress, err)
	}

	// register to grpc
	sExternal := grpc.NewServer()
	pb.RegisterKeyValueStoreServer(sExternal, newServer)

	sInternal := grpc.NewServer()
	pb.RegisterKeyValueStoreInternalServer(sInternal, newServer)

	// start store vectorclock every 1s
	go func() {
		for {
			newServer.StoreVectorClock()
			time.Sleep(time.Second)
		}
	}()

	// start serving
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
	log.Printf("=== Finished setting server %v ===\n", nodeId)

	// sleep forever
	select {}
}
