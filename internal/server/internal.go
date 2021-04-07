package server

import (
	"context"
	"github.com/heyuhang0/DSCProject/internal/nodemgr"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"github.com/syndtr/goleveldb/leveldb"
	dberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

// get replica issued from server responsible for the get operation
func (s *server) GetRep(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Getting replica for key: {%v} from local db\n", string(req.Key))
	s.vectorClock.MergeClock(vc.FromDTO(req.Vectorclock).Vclock)

	// add prefix to key to distinguish from persistent vector clock
	realKey := append([]byte("data_"), req.Key...)
	dataBytes, err := s.db.Get(realKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound || err == dberrors.ErrNotFound { // key not found
			return nil, status.Error(codes.NotFound, "LEVELDB KEY NOT FOUND")
		}
		return nil, err
	}

	data := &pb.VersionedData{}
	if err = proto.Unmarshal(dataBytes, data); err != nil {
		return nil, err
	}

	s.vectorClock.Advance()
	return &pb.GetRepResponse{Data: data, Vectorclock: vc.ToDTO(s.vectorClock)}, nil
}

// put replica issued from server responsible for the put operation
func (s *server) PutRep(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	log.Printf("Putting replica key: {%v}, val: {%v} to local db", string(req.Key), string(req.Data.Object))
	s.vectorClock.MergeClock(vc.FromDTO(req.Vectorclock).Vclock)

	dataBytes, err := proto.Marshal(req.Data)
	if err != nil {
		return nil, err
	}

	// add prefix to key to distinguish from persistent vector clock
	realKey := append([]byte("data_"), req.Key...)
	err = s.db.Put(realKey, dataBytes, nil)
	s.vectorClock.Advance()
	return &pb.PutRepResponse{Vectorclock: vc.ToDTO(s.vectorClock)}, err
}

func (s *server) SendHeartBeat() {
	seedServers := s.seedServerId
	nodeInfo, err := s.nodes.GetNodeInfo(s.id)
	if err != nil {
		panic("my info is not found in my node manager")
	}
	for _, serverId := range seedServers {
		go func(serverId uint64){
			if serverId == s.id {
				return
			}
			log.Printf("Tring to send hearbeat to seed server %v", serverId)
			heartBeatReq := pb.HeartBeatRequest{
				Id:              s.id,
				InternalAddress: nodeInfo.InternalAddress,
				ExternalAddress: nodeInfo.ExternalAddress,
				Version:         time.Now().UnixNano(),
			}
			seedServer, seedErr := s.nodes.GetInternalClient(serverId)
			ctxHB, cancel := context.WithTimeout(context.Background(), s.timeout)
			defer cancel()
			if seedErr == nil {
				resp, respErr := seedServer.HeartBeat(ctxHB, &heartBeatReq)
				// merge history if response is successful
				if respErr == nil && resp != nil {
					log.Printf("Received Heartbeat response from seed server %v", serverId)
					history := resp.Nodes
					s.nodes.ImportHistory(history)
				}else{
					log.Printf("Seed server %v is not contectable", serverId)
				}
			}else{
				log.Printf("Seed server %v is not contectable", serverId)
			}
		}(serverId)
	}
}

func (s *server) HeartBeat(ctx context.Context, req *pb.HeartBeatRequest) (*pb.HeartBeatResponse, error) {
	log.Printf("received heartbeat request from node %v with version %v, replying", req.Id, req.Version)
	// construct node info
	nodeInfo := nodemgr.NodeInfo{
		ID:              req.Id,
		Alive:           true,
		InternalAddress: req.InternalAddress,
		ExternalAddress: req.ExternalAddress,
		Version:         req.Version,
	}
	// update node info with new timeout
	s.nodes.UpdateNodeWithExpire(&nodeInfo, 10*time.Second)

	return &pb.HeartBeatResponse{
		Id:    s.id,
		Nodes: s.nodes.ExportHistory(),
	}, nil
}
