package server

import (
	"bytes"
	"context"
	"errors"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
	dberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	id                  int
	allServerID         []int
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
		id:          id,
		allServerID: allServerID,
		numReplica:  numReplica,
		numRead:     numRead,
		numWrite:    numWrite,
		timeout:     timeout,
		db:          db}
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
	keyString := string(req.Key)
	log.Printf("Received GET REQUEST from clients for key {%v}\n", keyString)

	// get preference list for the key
	preferenceList, errGetPref := s.GetMockPreferenceList(req.Key)
	if errGetPref != nil {
		return nil, errGetPref
	}

	// check if myself is in preference list
	if !contains(preferenceList, s.id) {
		// need to forward the message to the one in preference list, haven't implemented
		log.Printf("need to forward the get request to other server, I am not in preference list")
	}

	notifyChan := make(chan GetRepMessage, s.numReplica)
	// send getRep request to all servers in preference list
	for _, peerServerId := range preferenceList {
		reqRep := pb.GetRepRequest{Key: req.Key}
		// send getRep request to servers with id peerServerId
		go func(peerServerId int, notifyChan chan GetRepMessage) {
			// gRPC context
			clientDeadline := time.Now().Add(s.timeout)
			ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
			defer cancel()

			var dataRep *pb.GetRepResponse
			var err error
			// myself
			if peerServerId == s.id {
				dataRep, err = s.GetRep(ctxRep, &reqRep)
			} else { // other servers
				peerServerInterface, exist := s.otherServerInstance.Load(peerServerId)
				if !exist {
					panic("peer id is not stored in server " + strconv.Itoa(s.id))
				}
				peerServer, ok := peerServerInterface.(pb.KeyValueStoreInternalClient)
				if !ok {
					panic("peer server is not the correct type " + strconv.Itoa(s.id))
				}
				dataRep, err = peerServer.GetRep(ctxRep, &reqRep)
			}
			// notify main routine
			if err != nil {
				notifyChan <- GetRepMessage{peerServerId, err, nil}
			} else {
				notifyChan <- GetRepMessage{peerServerId, nil, dataRep.Object}
			}
		}(peerServerId, notifyChan)
	}

	var vals [][]byte
	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(preferenceList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received GET repica response from server %v for key {%v}: val {%v}", notifyMsg.id, keyString, string(notifyMsg.data))
				successCount++
				vals = append(vals, notifyMsg.data)
			} else if e, ok := status.FromError(notifyMsg.err); ok && e.Code() == codes.NotFound { // key not found
				log.Printf("Received GET repica response from server %v for key {%v}: key not found", notifyMsg.id, keyString)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received GET repica response from server %v for key {%v}: occured error: %v", notifyMsg.id, keyString, notifyMsg.err.Error())
			}
		//	timeout
		case <-time.After(s.timeout):
			break
		}
		if successCount >= s.numRead {
			break
		}
	}
	log.Printf("Finished GETTING replica for key {%v}, received {%v} response", keyString, successCount)
	// all server in preference list are down
	if successCount == 0 {
		return nil, errors.New(errorMsg)
	}
	// at least one successful response
	successStatus := pb.SuccessStatus_PARTIAL_SUCCESS
	// at least numRead requirements
	if successCount >= s.numRead {
		successStatus = pb.SuccessStatus_FULLY_SUCCESS
	}
	// all server does not have values: key does not exist in db
	if len(vals) == 0 {
		return &pb.GetResponse{Object: nil, SuccessStatus: successStatus, FoundKey: pb.FoundKey_KEY_NOT_FOUND}, nil
	}
	// check if all the element are same and return different things
	if allSame(vals) {
		return &pb.GetResponse{Object: vals[0], SuccessStatus: successStatus, FoundKey: pb.FoundKey_KEY_FOUND}, nil
	} else {
		// TODO: return the arbitrary latest value
		return &pb.GetResponse{Object: vals[0], SuccessStatus: successStatus, FoundKey: pb.FoundKey_KEY_FOUND}, nil
	}
}

// Put key issued from the client
func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	keyString := string(req.Key)
	valString := string(req.Object)
	log.Printf("Received PUT request from clients: key {%v}, val{%v}", keyString, valString)

	// get preference list for the key
	preferenceList, errGetPref := s.GetMockPreferenceList(req.Key)
	if errGetPref != nil {
		return nil, errGetPref
	}
	// check if myself is in preference list
	if !contains(preferenceList, s.id) {
		// need to forward the message to the one in preference list, haven't implemented
		log.Printf("need to forward the get request to other server, I am not in preference list")
	}

	// need numWrite to finish the operation
	notifyChan := make(chan PutRepMessage, s.numReplica)

	// send PutRep request to all servers in preference list
	for _, peerServerId := range preferenceList {
		reqRep := pb.PutRepRequest{Key: req.Key, Object: req.Object}
		var err error
		// send putRep request to servers with id peerServerId
		go func(peerServerId int, notifyChan chan PutRepMessage) {
			// gPRC context
			clientDeadline := time.Now().Add(s.timeout)
			ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
			defer cancel()
			// myself
			if peerServerId == s.id {
				_, err = s.PutRep(ctxRep, &reqRep)
			} else { // other server
				peerServerInterface, exist := s.otherServerInstance.Load(peerServerId)
				if !exist {
					panic("peer id is not stored in server " + strconv.Itoa(s.id))
				}
				peerServer, ok := peerServerInterface.(pb.KeyValueStoreInternalClient)
				if !ok {
					panic("peer server is not the correct type " + strconv.Itoa(s.id))
				}
				_, err = peerServer.PutRep(ctxRep, &reqRep)
			}
			notifyChan <- PutRepMessage{peerServerId, err}
		}(peerServerId, notifyChan)
	}

	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(preferenceList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received PUT repica response from server %v for key {%v} val {%v}: SUCCESS", notifyMsg.id, keyString, valString)
				successCount++
			} else {
				errorMsg = errorMsg + notifyMsg.err.Error() + ";"
				log.Printf("Received PUT repica response from server %v for key {%v} val {%v}: ERROR occured: %v\n", notifyMsg.id, keyString, valString, notifyMsg.err.Error())
			}
		case <-time.After(s.timeout):
			break
		}
		if successCount >= 3 {
			break
		}
	}

	log.Printf("Finished PUTTING replica for key {%v} val {%v}, received {%v} response", keyString, valString, successCount)
	// all servers failed
	if successCount == 0 {
		return nil, errors.New(errorMsg)
	}
	// meet numWrite requirements
	if successCount >= s.numWrite {
		return &pb.PutResponse{SuccessStatus: pb.SuccessStatus_FULLY_SUCCESS}, nil
	}
	return &pb.PutResponse{SuccessStatus: pb.SuccessStatus_PARTIAL_SUCCESS}, nil
}

// get replica issued from server responsible for the get operation
func (s *server) GetRep(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Getting replica for key: {%v} from local db\n", string(req.Key))
	data, err := s.db.Get(req.Key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound || err == dberrors.ErrNotFound { // key not found
			return nil, status.Error(codes.NotFound, "LEVELDB KEY NOT FOUND")
		}
		return nil, err
	}
	return &pb.GetRepResponse{Object: data}, nil
}

// put replica issued from server responsible for the put operation
func (s *server) PutRep(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	log.Printf("Putting replica key: {%v}, val: {%v} to local db", string(req.Key), string(req.Object))
	err := s.db.Put(req.Key, req.Object, nil)
	return &pb.PutRepResponse{}, err
}
