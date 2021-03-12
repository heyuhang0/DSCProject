package server

import (
	"context"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
)

type server struct {
	pb.UnimplementedKeyValueStoreServer
	db *leveldb.DB
}

func NewServer(db *leveldb.DB) *server {
	return &server{db: db}
}

func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	data, err := s.db.Get(req.Key, nil)
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Object: data}, nil
}

func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	err := s.db.Put(req.Key, req.Object, nil)
	return &pb.PutResponse{}, err
}
