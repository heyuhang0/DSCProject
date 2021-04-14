package server

import (
	"bytes"
	"encoding/gob"
	"github.com/heyuhang0/DSCProject/internal/keyedmutex"
	"github.com/heyuhang0/DSCProject/internal/nodemgr"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"time"
)

type server struct {
	pb.UnimplementedKeyValueStoreServer
	pb.UnimplementedKeyValueStoreInternalServer

	id           uint64
	seedServerId []uint64

	numRead    int
	numWrite   int
	numReplica int
	numVNodes  int
	timeout    time.Duration

	nodes       *nodemgr.Manager
	db          *leveldb.DB
	vectorClock *vc.VectorClock

	putMutex keyedmutex.KeyedMutex
}

// create a new server
func NewServer(id uint64, seedServerId []uint64, numReplica, numRead, numWrite, numVNodes int, timeout time.Duration, nodes *nodemgr.Manager, db *leveldb.DB) *server {
	vectorClock := vc.NewVectorClock(int(id))
	vcKey := []byte("vc")
	vcBytes, err := db.Get(vcKey, nil) // get vectorclock
	if err == nil {
		// meaning there exists vectorclock in the database
		buf := bytes.NewBuffer(vcBytes)
		dec := gob.NewDecoder(buf)
		v := make(map[int]int)
		dec.Decode(&v)
		vectorClock.Vclock = v
	}
	log.Println("Initialize the vector clock:", vectorClock.Vclock)
	return &server{
		id:           id,
		seedServerId: seedServerId,
		numReplica:   numReplica,
		numRead:      numRead,
		numWrite:     numWrite,
		numVNodes:    numVNodes,
		timeout:      timeout,
		nodes:        nodes,
		db:           db,
		vectorClock:  vectorClock,
		putMutex:     make(keyedmutex.KeyedMutex, 128),
	}
}

func (s *server) GetPreferenceList(key []byte) []uint64 {
	return s.nodes.GetPreferenceList(key, s.numReplica)
}

func (s *server) StoreVectorClock() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(s.vectorClock.GetVectorClock())
	vcKey := []byte("vc")
	_ = s.db.Put(vcKey, buf.Bytes(), nil)
}
