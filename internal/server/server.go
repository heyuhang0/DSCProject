package server

import (
	"bytes"
	"context"
	"errors"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
	dberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"log"
	"strconv"
	"sync"
	"time"
)

type GetRepMessage struct {
	id   int
	err  error
	data []byte
}

type PutRepMessage struct {
	id  int
	err error
}

type Consistent struct {
	consistentStructure []int
}

type server struct {
	pb.UnimplementedKeyValueStoreServer
	pb.UnimplementedKeyValueStoreInternalServer
	id          int
	allServerID []int
	//otherServerInstance []pb.KeyValueStoreInternalClient
	otherServerInstance *sync.Map
	Consistent          Consistent
	numDescendants      int
	numRead             int
	numWrite            int
	numReplica          int
	timeout             time.Duration
	db                  *leveldb.DB
}

func (s *server) SetOtherServerInstance(otherServerInstance *sync.Map) {
	s.otherServerInstance = otherServerInstance
}

func (s *server) SetConsistent(consistent Consistent) {
	s.Consistent = consistent
}

// create a new server
func NewServer(id int, allServerID []int, numReplica, numRead, numWrite int, timeout time.Duration, db *leveldb.DB) *server {
	return &server{
		id: id,
		allServerID: allServerID,
		numReplica: numReplica,
		numRead: numRead,
		numWrite: numWrite,
		timeout: timeout,
		db: db}
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
	log.Printf("Received GET request from clients for {key {%v}}\n", string(req.Key))
	preferenceList, errGetPref := s.GetMockPreferenceList(req.Key)
	if errGetPref != nil {
		return nil, errGetPref
	}

	// check if myself is in preference list
	if !contains(preferenceList, s.id) {
		// need to forward the message to the one in preference list, haven't implemented
		log.Println("need to forward the get request to other server, I am not in preference list")
	}

	notifyChan := make(chan GetRepMessage, s.numReplica)
	for _, peerServerId := range preferenceList {
		// if it is server it self
		if peerServerId == s.id {
			go func(notifyChan chan GetRepMessage) {
				log.Printf("Getting {key {%v}} from local db\n", string(req.Key))
				data, err := s.db.Get(req.Key, nil)
				notifyChan <- GetRepMessage{s.id, err, data}
			}(notifyChan)
			continue
		}
		peerServerInterface, exist := s.otherServerInstance.Load(peerServerId)
		if !exist {
			panic("peer id is not stored in server " + strconv.Itoa(s.id))
		}
		peerServer, ok := peerServerInterface.(pb.KeyValueStoreInternalClient)
		if !ok {
			panic("peer server is not the correct type " + strconv.Itoa(s.id))
		}
		go func(peerServerId int, peerServer pb.KeyValueStoreInternalClient, notifyChan chan GetRepMessage) {
			reqRep := pb.GetRepRequest{Key: req.Key}
			clientDeadline := time.Now().Add(s.timeout)
			ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
			defer cancel()
			dataRep, err := peerServer.GetRep(ctxRep, &reqRep)
			// notify main routine
			if err != nil {
				notifyChan <- GetRepMessage{peerServerId, err, nil}
			} else {
				notifyChan <- GetRepMessage{peerServerId, nil, dataRep.Object}
			}
		}(peerServerId, peerServer, notifyChan)
	}

	var vals [][]byte
	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(preferenceList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				successCount++
				vals = append(vals, notifyMsg.data)
			} else if notifyMsg.err == leveldb.ErrNotFound || notifyMsg.err == dberrors.ErrNotFound{
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("server %v occured error: %v", notifyMsg.id, notifyMsg.err.Error())
			}
		case <-time.After(s.timeout):
			break
		}
		if successCount >= 3 {
			if len(vals) == 0 {
				return &pb.GetResponse{Object: nil, SuccessStatus: pb.SuccessStatus_FULLY_SUCCESS, FoundKey: pb.FoundKey_KEY_NOT_FOUND}, nil
			}
			// check if all the element are same and return different things
			if allSame(vals) {
				return &pb.GetResponse{Object: vals[0], SuccessStatus: pb.SuccessStatus_FULLY_SUCCESS, FoundKey: pb.FoundKey_KEY_FOUND}, nil
			} else {
				// TODO: return the arbitrary latest value
				return &pb.GetResponse{Object: vals[0], SuccessStatus: pb.SuccessStatus_FULLY_SUCCESS, FoundKey: pb.FoundKey_KEY_FOUND}, nil
			}
		}
	}

	// all server in preference list are done
	if successCount == 0 {
		return nil, errors.New(errorMsg)
	}
	// all server does not have values
	if len(vals) == 0 {
		return &pb.GetResponse{Object: nil, SuccessStatus: pb.SuccessStatus_PARTIAL_SUCCESS, FoundKey: pb.FoundKey_KEY_NOT_FOUND}, nil
	}
	// check if all the element are same and return different things
	if allSame(vals) {
		return &pb.GetResponse{Object: vals[0], SuccessStatus: pb.SuccessStatus_PARTIAL_SUCCESS, FoundKey: pb.FoundKey_KEY_FOUND}, nil
	} else {
		// TODO: return the arbitrary latest value
		return &pb.GetResponse{Object: vals[0], SuccessStatus: pb.SuccessStatus_PARTIAL_SUCCESS, FoundKey: pb.FoundKey_KEY_FOUND}, nil
	}
}

// Put key issued from the client
func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	log.Println("Received PUT request from clients")
	preferenceList, errGetPref := s.GetMockPreferenceList(req.Key)
	if errGetPref != nil {
		return nil, errGetPref
	}
	// check if myself is in preference list
	if !contains(preferenceList, s.id) {
		// need to forward the message to the one in preference list, haven't implemented
		log.Println("need to forward the get request to other server, I am not in preference list")
	}

	// need numWrite to finish the operation
	notifyChan := make(chan PutRepMessage, s.numReplica)
	for _, peerServerId := range preferenceList {
		if peerServerId == s.id {
			go func() {
				err := s.db.Put(req.Key, req.Object, nil)
				log.Printf("Putting {key: {%v}, val: {%v}} to local db", string(req.Key), string(req.Object))
				notifyChan <- PutRepMessage{s.id, err}
			}()
			continue
		}
		peerServerInterface, exist := s.otherServerInstance.Load(peerServerId)
		if !exist {
			panic("peer id is not stored in server " + strconv.Itoa(s.id))
		}
		peerServer, ok := peerServerInterface.(pb.KeyValueStoreInternalClient)
		if !ok {
			panic("peer server is not the correct type " + strconv.Itoa(s.id))
		}
		go func(peerServerId int, peerServer pb.KeyValueStoreInternalClient) {
			reqRep := pb.PutRepRequest{Key: req.Key, Object: req.Object}
			clientDeadline := time.Now().Add(s.timeout)
			ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
			defer cancel()
			_, err := peerServer.PutRep(ctxRep, &reqRep)
			notifyChan <- PutRepMessage{peerServerId, err}
		}(peerServerId, peerServer)
	}

	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(preferenceList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				successCount++
			} else {
				errorMsg =  errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("server %v error occured: %v\n", notifyMsg.id, notifyMsg.err.Error())
			}
		case <-time.After(s.timeout):
			break
		}
		if successCount >= 3 {
			// check if all the element are same and return different things
			return &pb.PutResponse{SuccessStatus: pb.SuccessStatus_FULLY_SUCCESS}, nil
		}
	}
	if successCount == 0 {
		return nil, errors.New(errorMsg)
	}
	return &pb.PutResponse{SuccessStatus: pb.SuccessStatus_PARTIAL_SUCCESS}, nil
}

// get replica issued from server responsible for the get operation
func (s *server) GetRep(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Getting replica for {key: {%v}} from local db\n", string(req.Key))
	data, err := s.db.Get(req.Key, nil)
	if err != nil {
		return nil, err
	}
	return &pb.GetRepResponse{Object: data}, nil
}

// put replica issued from server responsible for the put operation
func (s *server) PutRep(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	log.Printf("Putting replica {key: {%v}, val: {%v}} to local db", string(req.Key), string(req.Object))
	err := s.db.Put(req.Key, req.Object, nil)
	return &pb.PutRepResponse{}, err
}
