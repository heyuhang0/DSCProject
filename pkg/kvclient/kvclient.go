package kvclient

import (
	"context"
	"encoding/json"
	"errors"
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
	ID      uint64
	Address string
}

type ClientConfig struct {
	NodeTimeoutMs int
	Retry         int
	SeedNodes     []*SeedNodeConfig
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

type KeyValueStoreClient struct {
	mu sync.RWMutex

	nodeTimeout time.Duration
	retry       int

	seeds      []uint64
	addresses  map[uint64]string
	connPool   map[string]pb.KeyValueStoreClient
	consistent *consistent.Consistent
}

func NewKeyValueStoreClient(config *ClientConfig) *KeyValueStoreClient {
	seeds := make([]uint64, len(config.SeedNodes))
	addresses := make(map[uint64]string)

	for i, seedNode := range config.SeedNodes {
		seeds[i] = seedNode.ID
		addresses[seedNode.ID] = seedNode.Address
	}

	client := &KeyValueStoreClient{
		nodeTimeout: time.Duration(config.NodeTimeoutMs) * time.Millisecond,
		retry:       config.Retry,
		seeds:       seeds,
		addresses:   addresses,
		connPool:    make(map[string]pb.KeyValueStoreClient),
	}

	// update consistent hashing in the background
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			_ = client.updateRing(ctx)
			<-ctx.Done()
			cancel()
		}
	}()

	return client
}

func (k *KeyValueStoreClient) getClient(nodeID uint64) (pb.KeyValueStoreClient, error) {
	k.mu.RLock()

	// Get address
	address, ok := k.addresses[nodeID]
	if !ok {
		k.mu.RUnlock()
		return nil, fmt.Errorf("NodeManager: nodeID {%v} not found", nodeID)
	}

	// Reuse the existing connection
	if client, ok := k.connPool[address]; ok {
		k.mu.RUnlock()
		return client, nil
	}

	// Create a new connection
	k.mu.RUnlock()
	k.mu.Lock()
	defer k.mu.Unlock()

	if client, ok := k.connPool[address]; ok {
		return client, nil
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewKeyValueStoreClient(conn)
	k.connPool[address] = client
	return client, nil
}

func (k *KeyValueStoreClient) retryForEveryNode(ctx context.Context, nodes []uint64, do func(ctx context.Context, client pb.KeyValueStoreClient) error) error {
	if len(nodes) == 0 {
		return errors.New("no nodes available")
	}
	errorMessages := make([]string, 0)
	for _, nodeID := range nodes {
		client, err := k.getClient(nodeID)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			if ctx.Err() != nil {
				break
			}
			continue
		}

		nodeCtx, cancel := context.WithTimeout(ctx, k.nodeTimeout)
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

func (k *KeyValueStoreClient) updateRing(ctx context.Context) error {
	seedIndex := -1
	var resp *pb.GetRingResponse

	err := k.retryForEveryNode(ctx, k.seeds, func(ctx context.Context, client pb.KeyValueStoreClient) (err error) {
		seedIndex += 1
		resp, err = client.GetRing(ctx, &pb.GetRingRequest{})
		return
	})
	if err != nil {
		return err
	}

	// swap seeds order, so that next time can get from the alive seed server directly
	if seedIndex != 0 {
		k.seeds[0], k.seeds[seedIndex] = k.seeds[seedIndex], k.seeds[0]
	}

	// update consistent hashing
	k.mu.Lock()
	if k.consistent == nil {
		k.consistent = consistent.NewConsistent(int(resp.NumVNodes))
	}
	for _, nodeInfo := range resp.Nodes {
		k.addresses[nodeInfo.Id] = nodeInfo.ExternalAddress
		if nodeInfo.Alive {
			k.consistent.AddNode(nodeInfo.Id)
		} else {
			k.consistent.RemoveNode(nodeInfo.Id)
		}
	}
	k.mu.Unlock()

	return err
}

func (k *KeyValueStoreClient) retryForKey(ctx context.Context, key []byte, do func(ctx context.Context, client pb.KeyValueStoreClient) error) error {
	k.mu.RLock()
	if k.consistent == nil {
		k.mu.RUnlock()
		err := k.updateRing(ctx)
		if err != nil {
			return err
		}
		k.mu.RLock()
	}
	preferenceList := k.consistent.GetNodes(key, k.retry)
	k.mu.RUnlock()
	return k.retryForEveryNode(ctx, preferenceList, do)
}

func (k *KeyValueStoreClient) Put(ctx context.Context, in *pb.PutRequest, opts ...grpc.CallOption) (response *pb.PutResponse, err error) {
	err = k.retryForKey(ctx, in.Key, func(nodeCtx context.Context, client pb.KeyValueStoreClient) error {
		response, err = client.Put(nodeCtx, in, opts...)
		return err
	})
	return
}

func (k *KeyValueStoreClient) Get(ctx context.Context, in *pb.GetRequest, opts ...grpc.CallOption) (response *pb.GetResponse, err error) {
	err = k.retryForKey(ctx, in.Key, func(nodeCtx context.Context, client pb.KeyValueStoreClient) error {
		response, err = client.Get(nodeCtx, in, opts...)
		return err
	})
	return
}
