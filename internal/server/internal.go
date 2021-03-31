package server

import (
	"context"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/heyuhang0/DSCProject/pkg/vc"
	"github.com/syndtr/goleveldb/leveldb"
	dberrors "github.com/syndtr/goleveldb/leveldb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"log"
)

// get replica issued from server responsible for the get operation
func (s *server) GetRep(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Printf("Getting replica for key: {%v} from local db\n", string(req.Key))
	s.vectorClock.MergeClock(vc.FromDTO(req.Vectorclock).Vclock)

	dataBytes, err := s.db.Get(req.Key, nil)
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

	err = s.db.Put(req.Key, dataBytes, nil)
	s.vectorClock.Advance()
	return &pb.PutRepResponse{Vectorclock: vc.ToDTO(s.vectorClock)}, err
}
