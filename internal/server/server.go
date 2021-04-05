package server

import (
	"bytes"
	"encoding/gob"
	"github.com/heyuhang0/DSCProject/internal/nodemgr"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
)

type server struct {
	pb.UnimplementedKeyValueStoreServer
	pb.UnimplementedKeyValueStoreInternalServer

	id uint64

	numRead    int
	numWrite   int
	numReplica int
	numVNodes  int
	timeout    time.Duration

	nodes       *nodemgr.Manager
	db          *leveldb.DB
	vectorClock *vc.VectorClock
}

// create a new server
func NewServer(id uint64, numReplica, numRead, numWrite, numVNodes int, timeout time.Duration, nodes *nodemgr.Manager, db *leveldb.DB) *server {
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
	return &server{
		id:          id,
		numReplica:  numReplica,
		numRead:     numRead,
		numWrite:    numWrite,
		numVNodes:   numVNodes,
		timeout:     timeout,
		nodes:       nodes,
		db:          db,
		vectorClock: vectorClock,
	}
}

func (s *server) GetPreferenceList(key []byte) []uint64 {
	return s.nodes.GetPreferenceList(key, s.numReplica)
}

func (s *server) GetVectorClock() map[int]int {
	return s.vectorClock.Vclock
}

func (s *server) StoreVectorClock() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(s.vectorClock.Vclock)
	vcKey := []byte("vc")
	_ = s.db.Put(vcKey, buf.Bytes(), nil)
}
