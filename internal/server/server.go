package server

import (
	"bytes"
	"context"
	"errors"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"strconv"
	"sync"
)

type Consistent struct {
	consistentStructure []int
}

type server struct {
	pb.UnimplementedKeyValueStoreServer
	pb.UnimplementedKeyValueStoreInternalServer
	id          int
	allServerID []int
	//otherServerInstance []pb.KeyValueStoreInternalClient
	otherServerInstance map[int]pb.KeyValueStoreInternalClient
	Consistent          Consistent
	numDescendants      int
	numRead             int
	numWrite            int
	numReplica          int
	db                  *leveldb.DB
	vectorClock         *vc.VectorClock
}

func (s *server) SetOtherServerInstance(otherServerInstance map[int]pb.KeyValueStoreInternalClient) {
	s.otherServerInstance = otherServerInstance
}

func (s *server) SetConsistent(consistent Consistent) {
	s.Consistent = consistent
}

// create a new server
func NewServer(id int, allServerID []int, numReplica, numRead, numWrite int, db *leveldb.DB) *server {
	return &server{id: id, allServerID: allServerID, numReplica: numReplica, numRead: numRead, numWrite: numWrite, db: db, vectorClock: vc.NewVectorClock(id)}
}

// contains
func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// the index of machine with id {id}
func indexOf(id int, data []int) int {
	for k, v := range data {
		if id == v {
			return k
		}
	}
	return -1
}

// check whether all byte array in an array are the same
func allSame(data [][]byte) bool {
	for i, _ := range data {
		if bytes.Compare(data[0], data[i]) != 0 {
			return false
		}
	}
	return true
}

// not used any more
func (s *server) GetDescendants() ([]pb.KeyValueStoreInternalClient, error) {
	numPeerReplica := s.numReplica
	if numPeerReplica > len(s.allServerID) {
		return nil, errors.New("descendents number bigger than total machine number")
	}
	var descendants []pb.KeyValueStoreInternalClient
	idIndex := indexOf(s.id, s.allServerID)
	for i := 1; i <= numPeerReplica; i++ {
		index := (idIndex + i) % len(s.allServerID)
		if index < idIndex {
			descendants = append(descendants, s.otherServerInstance[index])
		} else if index > idIndex {
			descendants = append(descendants, s.otherServerInstance[index-1])
		}
	}
	return descendants, nil
}

func (s *server) GetMockPreferenceList(key []byte) ([]int, error) {
	numPeerReplica := s.numReplica
	if numPeerReplica > len(s.allServerID) {
		return nil, errors.New("replica number bigger than total machine number")
	}
	var preferenceList []int
	idIndex := indexOf(s.id, s.allServerID)
	for i := 0; i < numPeerReplica; i++ {
		index := (idIndex + i) % len(s.allServerID)
		preferenceList = append(preferenceList, s.allServerID[index])
	}
	return preferenceList, nil
}

// get key, issued from client
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Println("Received GET request from clients")
	preferenceList, err := s.GetMockPreferenceList(req.Key)
	if err != nil {
		return nil, err
	}

	// check if myself is in preference list
	if !contains(preferenceList, s.id) {
		// need to forward the message to the one in preference list, haven't implemented
		log.Println("need to forward the get request to other server, I am not in preference list")
	}

	// need numRead to finish this operation
	var wg sync.WaitGroup
	wg.Add(s.numReplica)
	var vals [][]byte

	for _, peerServerId := range preferenceList {
		// if it is server it self
		if peerServerId == s.id {
			go func() {
				defer wg.Done()
				data, err := s.db.Get(req.Key, nil)
				if err == nil {
					vals = append(vals, data)
				}
			}()
			continue
		}
		peerServer, exist := s.otherServerInstance[peerServerId]
		if !exist {
			panic("peer id is not stored in server " + strconv.Itoa(s.id))
		}
		go func(peerServer pb.KeyValueStoreInternalClient) {
			defer wg.Done()
			s.vectorClock.Advance()
			reqRep := pb.GetRepRequest{Key: req.Key, Vectorclock: vc.ToDTO(s.vectorClock)}
			dataRep, errRoutine := peerServer.GetRep(ctx, &reqRep)
			// need to merge vector clock
			if dataRep != nil {
				// if no data in the node, then skip
				s.vectorClock.MergeClock(vc.FromDTO(dataRep.Vectorclock).Vclock)
			}
			// skip nil value for now
			if errRoutine != nil {
				return
			}
			log.Println("Received replica from peer server")
			vals = append(vals, dataRep.Object)
		}(peerServer)
	}

	// wait for {numRead} reads to finish
	wg.Wait()

	if len(vals) == 0 {
		return nil, errors.New("key not found")
	}

	// check if all the element are same and return different things
	if allSame(vals) {
		return &pb.GetResponse{Object: vals[0]}, nil
	} else {
		return &pb.GetResponse{Object: vals[0]}, nil
	}
}

// Put key issued from the client
func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	log.Println("Received PUT request from clients")
	preferenceList, err := s.GetMockPreferenceList(req.Key)
	if err != nil {
		return nil, err
	}
	// check if myself is in preference list
	if !contains(preferenceList, s.id) {
		// need to forward the message to the one in preference list, haven't implemented
		log.Println("need to forward the get request to other server, I am not in preference list")
	}

	// need numWrite to finish the operation
	var wg sync.WaitGroup
	wg.Add(s.numReplica)

	for _, peerServerId := range preferenceList {
		if peerServerId == s.id {
			go func() {
				defer wg.Done()
				err = s.db.Put(req.Key, req.Object, nil)
				log.Println("finish put local")
			}()
			continue
		}
		peerServer, exist := s.otherServerInstance[peerServerId]
		if !exist {
			panic("peer id is not stored in server " + strconv.Itoa(s.id))
		}
		go func(peerServer pb.KeyValueStoreInternalClient) {
			defer wg.Done()
			s.vectorClock.Advance()
			reqRep := pb.PutRepRequest{Key: req.Key, Object: req.Object, Vectorclock: vc.ToDTO(s.vectorClock)}
			repRes, errRoutine := peerServer.PutRep(context.Background(), &reqRep)
			// need to merge
			s.vectorClock.MergeClock(vc.FromDTO(repRes.Vectorclock).Vclock)
			log.Printf("finish put remote %v error %v", peerServer, err)
			if errRoutine != nil {
				err = errRoutine
			}
		}(peerServer)
	}
	// wait for {numWrite} reads to finish
	wg.Wait()

	return &pb.PutResponse{}, err
}

// get replica issued from server responsible for the get operation
func (s *server) GetRep(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Println("getting replica")
	s.vectorClock.MergeClock(vc.FromDTO(req.Vectorclock).Vclock)
	data, err := s.db.Get(req.Key, nil)
	if err != nil {
		return nil, err
	}
	s.vectorClock.Advance()
	return &pb.GetRepResponse{Object: data, Vectorclock: vc.ToDTO(s.vectorClock)}, nil
}

// put replica issued from server responsible for the put operation
func (s *server) PutRep(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	log.Println("putting replica")
	s.vectorClock.MergeClock(vc.FromDTO(req.Vectorclock).Vclock)
	err := s.db.Put(req.Key, req.Object, nil)
	s.vectorClock.Advance()
	return &pb.PutRepResponse{Vectorclock: vc.ToDTO(s.vectorClock)}, err
}
