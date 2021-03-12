package server

import (
	"bytes"
	"context"
	"errors"
	"log"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/syndtr/goleveldb/leveldb"
)

type server struct {
	pb.UnimplementedKeyValueStoreServer
	pb.UnimplementedKeyValueStoreInternalServer
	id                  int
	allServerID         []int
	otherServerInstance []pb.KeyValueStoreInternalClient
	numDescendants      int
	db                  *leveldb.DB
}

func NewServer(id int, allServerID []int, allServerInstance []pb.KeyValueStoreInternalClient, numDescendants int, db *leveldb.DB) *server {
	return &server{id:id, allServerID: allServerID, otherServerInstance:allServerInstance, numDescendants: numDescendants, db: db}
}

func indexOf(id int, data []int) int {
	for k, v := range data {
		if id == v {
			return k
		}
	}
	return -1
}

func allSame(data [][]byte) bool{
	for i, _ := range data {
		if bytes.Compare(data[0], data[i]) != 0{
			return false
		}
	}
	return true
}

func (s *server) GetDescendants() ([]pb.KeyValueStoreInternalClient, error) {
	numDes := s.numDescendants
	if numDes > len(s.allServerID) {
		return nil, errors.New("descendents number bigger than total machine number")
	}
	var descendants []pb.KeyValueStoreInternalClient
	idIndex := indexOf(s.id, s.allServerID)
	for i := 1 ; i<= numDes; i++ {
		index := (idIndex + i) % len(s.allServerID)
		if index < idIndex {
			descendants = append(descendants, s.otherServerInstance[index])
		}else if index > idIndex {
			descendants = append(descendants, s.otherServerInstance[index - 1])
		}
	}
	return descendants, nil
}

func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Println("Received GET request from clients")
	descendants, err := s.GetDescendants()
	if err != nil {
		return nil, err
	}

	var vals [][]byte
	data, err := s.db.Get(req.Key, nil)
	if err != nil {
		return nil, err
	}
	vals = append(vals, data)

	for _, v := range descendants {
		reqRep := pb.GetRepRequest{Key: req.Key}
		dataRep, err := v.GetRep(ctx, &reqRep)
		log.Println("Received replica from peer server")
		vals = append(vals, dataRep.Object)
		if err != nil {
			return nil, err
		}
	}

	// check if all the element are same and return different things
	if allSame(vals) {
		return &pb.GetResponse{Object: data}, nil
	}else {
		return &pb.GetResponse{Object: data}, nil
	}
}

//func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
//	data, err := s.db.Get(req.Key, nil)
//	if err != nil {
//		return nil, err
//	}
//	return &pb.GetResponse{Object: data}, nil
//}

func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	log.Println("Received PUT request from clients")
	descendants, err := s.GetDescendants()
	if err != nil {
		return nil, err
	}
	err = s.db.Put(req.Key, req.Object, nil)
	for _, v := range descendants {
		reqRep := pb.PutRepRequest{Key: req.Key, Object: req.Object}
		_, err := v.PutRep(ctx, &reqRep)
		if err != nil {
			return nil, err
		}
	}
	return &pb.PutResponse{}, err
}

//func (s *server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
//	err := s.db.Put(req.Key, req.Object, nil)
//	return &pb.PutResponse{}, err
//}

func (s *server) GetRep(ctx context.Context, req *pb.GetRepRequest) (*pb.GetRepResponse, error) {
	log.Println("getting replica")
	data, err := s.db.Get(req.Key, nil)
	if err != nil {
		return nil, err
	}
	return &pb.GetRepResponse{Object: data}, nil
}

func (s *server) PutRep(ctx context.Context, req *pb.PutRepRequest) (*pb.PutRepResponse, error) {
	log.Println("putting replica")
	err := s.db.Put(req.Key, req.Object, nil)
	return &pb.PutRepResponse{}, err
}
