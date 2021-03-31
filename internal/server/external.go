package server

import (
	"bytes"
	"context"
	"errors"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
	"time"
)

type GetRepMessage struct {
	id   int
	err  error
	data *pb.VersionedData
}

type PutRepMessage struct {
	id  int
	err error
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

// check whether all byte array in an array are the same
func allSame(data []*pb.VersionedData) bool {
	for i, _ := range data {
		if bytes.Compare(data[0].Object, data[i].Object) != 0 {
			return false
		}
	}
	return true
}

func getLatest(dataSlice []*pb.VersionedData) *pb.VersionedData {
	if len(dataSlice) == 0 {
		return nil
	}

	latest := dataSlice[0]
	for _, data := range dataSlice {
		// incoming version is newer iff
		// 1. every element in the old version exists in the incoming version, and
		// 2. the value of the incoming version >= old version
		// 3. at least one value > old version
		latestVersion := latest.Version.Vclock
		currVersion := data.Version.Vclock
		newer := false

		for i, val := range latestVersion {
			comingVal, ok := currVersion[i]
			if !ok || comingVal < val {
				newer = false
				break
			} else if comingVal > val {
				newer = true
			}
		}
		if newer {
			latest = data
		}
	}
	return latest
}

// get key, issued from client
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	keyString := string(req.Key)
	log.Printf("Received GET REQUEST from clients for key {%v}\n", keyString)

	// get preference list for the key
	preferenceList, errGetPref := s.GetPreferenceList(req.Key)
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
		reqRep := pb.GetRepRequest{
			Key:         req.Key,
			Vectorclock: vc.ToDTO(s.vectorClock),
		}
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

				// need to merge vector clock
				if err == nil && dataRep != nil {
					// if no data in the node, then skip
					s.vectorClock.MergeClock(vc.FromDTO(dataRep.Vectorclock).Vclock)
				}
			}
			// notify main routine
			if err != nil {
				notifyChan <- GetRepMessage{peerServerId, err, nil}
			} else {
				notifyChan <- GetRepMessage{peerServerId, nil, dataRep.Data}
			}
		}(peerServerId, notifyChan)
	}

	var vals []*pb.VersionedData
	successCount := 0
	errorMsg := "ERROR MESSAGE: "
	for i := 0; i < len(preferenceList); i++ {
		select {
		case notifyMsg := <-notifyChan:
			if notifyMsg.err == nil {
				log.Printf("Received GET repica response from server %v for key {%v}: val {%v}", notifyMsg.id, keyString, string(notifyMsg.data.Object))
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
		return &pb.GetResponse{Object: vals[0].Object, SuccessStatus: successStatus, FoundKey: pb.FoundKey_KEY_FOUND}, nil
	} else {
		log.Printf("Received inconsistent values for key {%v}: %v", keyString, vals)
		latest := getLatest(vals)
		log.Printf("Selected the lastest value {%v} for key {%v}", latest, keyString)
		return &pb.GetResponse{Object: latest.Object, SuccessStatus: successStatus, FoundKey: pb.FoundKey_KEY_FOUND}, nil
	}
}

// Put key issued from the client
func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	keyString := string(req.Key)
	valString := string(req.Object)
	log.Printf("Received PUT request from clients: key {%v}, val{%v}", keyString, valString)

	// get preference list for the key
	preferenceList, errGetPref := s.GetPreferenceList(req.Key)
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
	timestamp := vc.ToDTO(s.vectorClock)
	reqRep := &pb.PutRepRequest{
		Key: req.Key,
		Data: &pb.VersionedData{
			Object:  req.Object,
			Version: timestamp,
		},
		Vectorclock: timestamp,
	}
	for _, peerServerId := range preferenceList {
		var err error
		// send putRep request to servers with id peerServerId
		go func(peerServerId int, notifyChan chan PutRepMessage) {
			// gPRC context
			clientDeadline := time.Now().Add(s.timeout)
			ctxRep, cancel := context.WithDeadline(ctx, clientDeadline)
			defer cancel()
			// myself
			if peerServerId == s.id {
				_, err = s.PutRep(ctxRep, reqRep)
			} else { // other server
				peerServerInterface, exist := s.otherServerInstance.Load(peerServerId)
				if !exist {
					panic("peer id is not stored in server " + strconv.Itoa(s.id))
				}
				peerServer, ok := peerServerInterface.(pb.KeyValueStoreInternalClient)
				if !ok {
					panic("peer server is not the correct type " + strconv.Itoa(s.id))
				}
				var resp *pb.PutRepResponse
				resp, err = peerServer.PutRep(ctxRep, reqRep)

				// need to merge vector clock
				if err == nil && resp != nil {
					// if no data in the node, then skip
					s.vectorClock.MergeClock(vc.FromDTO(resp.Vectorclock).Vclock)
				}
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
