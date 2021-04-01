package kvclient

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/heyuhang0/DSCProject/pkg/consistent"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"google.golang.org/grpc"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type SeedNodeConfig struct {
	ID      int
	Address string
}

type ClientConfig struct {
	NumVirtualNodes int
	NodeTimeoutMs   int
	Retry           int
	SeedNodes       []*SeedNodeConfig
}

func NewClientConfigFromFile(configPath string) (*ClientConfig, error) {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = jsonFile.Close() }()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var config ClientConfig
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

type node struct {
	mu      sync.RWMutex
	address string
	client  pb.KeyValueStoreClient
}

func (s *node) getClient(ctx context.Context) (pb.KeyValueStoreClient, error) {
	s.mu.RLock()
	if s.client != nil {
		defer s.mu.RUnlock()
		return s.client, nil
	}
	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	// check again, since other machine may have created it
	// when we acquire write lock
	if s.client != nil {
		return s.client, nil
	}
	conn, err := grpc.DialContext(ctx, s.address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	s.client = pb.NewKeyValueStoreClient(conn)
	return s.client, nil
}

type KeyValueStoreClient struct {
	nodeTimeout time.Duration
	nodes       map[int]*node
	consistent  *consistent.Consistent
}

func NewKeyValueStoreClient(config *ClientConfig) *KeyValueStoreClient {
	// TODO initialize hash ring from seed node once implemented
	hashRing := consistent.NewConsistent(config.NumVirtualNodes)
	nodes := make(map[int]*node)

	for _, seedNode := range config.SeedNodes {
		hashRing.AddNode(uint64(seedNode.ID))
		nodes[seedNode.ID] = &node{
			address: seedNode.Address,
		}
	}

	return &KeyValueStoreClient{
		nodeTimeout: time.Duration(config.NodeTimeoutMs) * time.Millisecond,
		nodes:       nodes,
		consistent:  hashRing,
	}
}

func (k *KeyValueStoreClient) retryForEveryNode(ctx context.Context, key []byte, do func(ctx context.Context, client pb.KeyValueStoreClient) error) error {
	preferenceList := k.consistent.GetNodes(key, len(k.nodes))

	errorMessages := make([]string, 0)
	for _, nodeID := range preferenceList {
		nodeCtx, cancel := context.WithTimeout(ctx, k.nodeTimeout)

		client, err := k.nodes[int(nodeID)].getClient(nodeCtx)
		if err != nil {
			cancel()
			errorMessages = append(errorMessages, err.Error())
			if ctx.Err() != nil {
				break
			}
			continue
		}

		err = do(nodeCtx, client)
		cancel()
		if err == nil {
			return nil
		}

		errorMessages = append(errorMessages, err.Error())
		if ctx.Err() != nil {
			break
		}
	}
	return fmt.Errorf("multiple errors encountered: %v", strings.Join(errorMessages, "; "))
}

func (k *KeyValueStoreClient) Put(ctx context.Context, in *pb.PutRequest, opts ...grpc.CallOption) (response *pb.PutResponse, err error) {
	err = k.retryForEveryNode(ctx, in.Key, func(nodeCtx context.Context, client pb.KeyValueStoreClient) error {
		response, err = client.Put(nodeCtx, in, opts...)
		return err
	})
	return
}

func (k *KeyValueStoreClient) Get(ctx context.Context, in *pb.GetRequest, opts ...grpc.CallOption) (response *pb.GetResponse, err error) {
	err = k.retryForEveryNode(ctx, in.Key, func(nodeCtx context.Context, client pb.KeyValueStoreClient) error {
		response, err = client.Get(nodeCtx, in, opts...)
		return err
	})
	return
}
