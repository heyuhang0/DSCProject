package server

import (
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

	numRead        int
	numWrite       int
	numReplica     int
	timeout        time.Duration

	nodes       *nodemgr.Manager
	db          *leveldb.DB
	vectorClock *vc.VectorClock
}

// create a new server
func NewServer(id uint64, numReplica, numRead, numWrite int, timeout time.Duration, nodes *nodemgr.Manager, db *leveldb.DB) *server {
	return &server{
		id:          id,
		numReplica:  numReplica,
		numRead:     numRead,
		numWrite:    numWrite,
		timeout:     timeout,
		nodes:       nodes,
		db:          db,
		vectorClock: vc.NewVectorClock(int(id)),
	}
}

func (s *server) GetPreferenceList(key []byte) []uint64 {
	return s.nodes.GetPreferenceList(key, s.numReplica)
}
