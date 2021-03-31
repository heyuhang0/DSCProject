package server

import (
	"errors"
	"github.com/heyuhang0/DSCProject/pkg/consistent"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"time"
)

type server struct {
	pb.UnimplementedKeyValueStoreServer
	pb.UnimplementedKeyValueStoreInternalServer
	id                  int
	allServerID         []int
	otherServerInstance *sync.Map
	consistent          *consistent.Consistent
	numDescendants      int
	numRead             int
	numWrite            int
	numReplica          int
	timeout             time.Duration
	db                  *leveldb.DB
	vectorClock         *vc.VectorClock
}

// create a new server
func NewServer(id int, allServerID []int, numReplica, numRead, numWrite, numVNodes int, timeout time.Duration, db *leveldb.DB) *server {
	hashRing := consistent.NewConsistent(numVNodes)
	for _, serverID := range allServerID {
		hashRing.AddNode(uint64(serverID))
	}

	return &server{
		id:          id,
		allServerID: allServerID,
		numReplica:  numReplica,
		numRead:     numRead,
		numWrite:    numWrite,
		consistent:  hashRing,
		timeout:     timeout,
		db:          db,
		vectorClock: vc.NewVectorClock(id),
	}
}

func (s *server) SetOtherServerInstance(otherServerInstance *sync.Map) {
	s.otherServerInstance = otherServerInstance
}

func (s *server) GetPreferenceList(key []byte) ([]int, error) {
	numPeerReplica := s.numReplica
	if numPeerReplica > len(s.allServerID) {
		return nil, errors.New("replica number bigger than total machine number")
	}
	var preferenceList []int
	for _, serverID := range s.consistent.GetNodes(key, numPeerReplica) {
		preferenceList = append(preferenceList, int(serverID))
	}
	return preferenceList, nil
}
