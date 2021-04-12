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
	NumReplica      int
	NumRead         int
	NumWrite        int
	NumVirtualNodes int
	Timeout         int
	SeedServers     []*ServerConfig
}

func main() {
	// parse arguments
	serverConfig := flag.String("config", "./configs/default_config.json", "config file path")
	// seed server
	ifSeedSever := flag.Bool("seed", false, "whether the node is seed server")
	serverIdx := flag.Int("index", 1, "server index")
	// normal server
	nodeIdNormal := flag.Uint64("id", 0, "id of the server")
	internalAddressNormal := flag.String("internalAddr", "", "internal address of the server")
	externalAddressNormal := flag.String("externalAddr", "", "external address of the server")

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

	numRead := config.NumRead
	numWrite := config.NumWrite
	numReplica := config.NumReplica
	numVNodes := config.NumVirtualNodes
	timeout := time.Millisecond * time.Duration(config.Timeout)

	seedServers := config.SeedServers
	seedServerIds := make([]uint64, len(seedServers))
	for i, seedServer := range seedServers {
		seedServerIds[i] = seedServer.Id
	}

	log.Printf("id: %v, internalAddr: %v, externalAddr: %v", *nodeIdNormal, *internalAddressNormal, *externalAddressNormal)
	log.Printf("seedserver %v", *ifSeedSever)
	var nodeId uint64
	var internalAddress string
	var externalAddress string
	if !*ifSeedSever {
		nodeId = *nodeIdNormal
		internalAddress = *internalAddressNormal
		externalAddress = *externalAddressNormal
		if nodeId == 0 || internalAddress == "" || externalAddress == "" {
			log.Fatal("Please provide correct argument for normal server")
		}
	}
	if *ifSeedSever {
		if *serverIdx <= 0 || *serverIdx > len(seedServers) {
			log.Fatalf("Server index %v out of range [1, %v]", *serverIdx, len(seedServers))
		}
		localServer := seedServers[*serverIdx-1]
		nodeId = localServer.Id
		internalAddress = localServer.IpInternal + ":" + strconv.Itoa(localServer.PortInternal)
		externalAddress = localServer.IpExternal + ":" + strconv.Itoa(localServer.PortExternal)
	}

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
	for _, nodeConfig := range seedServers {
		alive := false
		if nodeConfig.Id == nodeId {
			alive = true
		}
		nodeManager.UpdateNode(&nodemgr.NodeInfo{
			ID:              nodeConfig.Id,
			Alive:           alive,
			InternalAddress: fmt.Sprintf("%v:%v", nodeConfig.IpInternal, nodeConfig.PortInternal),
			ExternalAddress: fmt.Sprintf("%v:%v", nodeConfig.IpExternal, nodeConfig.PortExternal),
			Version:         time.Now().UnixNano(),
		})
	}
	if !*ifSeedSever {
		nodeManager.UpdateNode(&nodemgr.NodeInfo{
			ID:              nodeId,
			Alive:           true,
			InternalAddress: internalAddress,
			ExternalAddress: externalAddress,
			Version:         time.Now().UnixNano(),
		})
	}
	ringVisualAddr := fmt.Sprintf("127.0.0.1:%d", 8000+*serverIdx)
	go func() {
		log.Fatal(nodeManager.ServeDashboard(ringVisualAddr))
	}()
	log.Printf("View consistent hashing ring on http://%v/", ringVisualAddr)

	// create server instance
	newServer := server.NewServer(nodeId, seedServerIds, numReplica, numRead, numWrite, numVNodes, timeout, nodeManager, db)

	// listen to external and internal ports
	lisInternal, err := net.Listen("tcp", internalAddress)
	log.Printf("Listening internal address: %v\n", internalAddress)
	if err != nil {
		log.Fatalf("Failed to listen to internal address %v: %v", internalAddress, err)
	}

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

	go func() {
		for {
			newServer.SendHeartBeat()
			time.Sleep(5 * time.Second)
		}
	}()
	// sleep forever
	select {}
}
