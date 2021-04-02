package nodemgr

import (
	"fmt"
	"github.com/heyuhang0/DSCProject/pkg/consistent"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

type NodeInfo struct {
	ID              uint64
	Alive           bool
	InternalAddress string
	ExternalAddress string
	Version         int64
}

func (n *NodeInfo) Copy() *NodeInfo {
	copied := NodeInfo{}
	copied = *n
	return &copied
}

func (n *NodeInfo) ToDTO() *pb.NodeInfo {
	return &pb.NodeInfo{
		Id:              n.ID,
		Alive:           n.Alive,
		InternalAddress: n.InternalAddress,
		ExternalAddress: n.ExternalAddress,
		Version:         n.Version,
	}
}

type NodeHistory map[uint64]*NodeInfo

func (n NodeHistory) Copy() NodeHistory {
	copied := make(NodeHistory)
	for key, value := range n {
		copied[key] = value.Copy()
	}
	return copied
}

func (n NodeHistory) ToDTO() map[uint64]*pb.NodeInfo {
	result := make(map[uint64]*pb.NodeInfo)
	for key, value := range n {
		result[key] = value.ToDTO()
	}
	return result
}

func NodeHistoryFromDTO(in map[uint64]*pb.NodeInfo) NodeHistory {
	result := make(NodeHistory)
	for key, value := range in {
		result[key] = &NodeInfo{
			ID:              value.Id,
			Alive:           value.Alive,
			InternalAddress: value.InternalAddress,
			ExternalAddress: value.ExternalAddress,
			Version:         value.Version,
		}
	}
	return result
}

type Manager struct {
	mu sync.RWMutex

	nodes        NodeHistory
	numVNodes    int
	consistent   *consistent.Consistent
	internalPool map[string]pb.KeyValueStoreInternalClient
	timers       map[uint64]*time.Timer
}

func NewManager(numVNodes int) *Manager {
	return &Manager{
		nodes:        make(NodeHistory),
		numVNodes:    numVNodes,
		consistent:   consistent.NewConsistent(numVNodes),
		internalPool: make(map[string]pb.KeyValueStoreInternalClient),
		timers:       make(map[uint64]*time.Timer),
	}
}

func (m *Manager) doUpdateNode(node *NodeInfo) {
	node = node.Copy()

	prevAlive := false
	if prev, ok := m.nodes[node.ID]; ok {
		prevAlive = prev.Alive
	}
	m.nodes[node.ID] = node

	if !prevAlive && node.Alive {
		log.Printf("Node %v has been added to consistent hash ring", node.ID)
		m.consistent.AddNode(node.ID)
	} else if prevAlive && !node.Alive {
		log.Printf("Node %v has been removed from consistent hash ring", node.ID)
		m.consistent.RemoveNode(node.ID)
	}
}

func (m *Manager) UpdateNode(node *NodeInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.doUpdateNode(node)

	// cancel expiration if there is any
	if timer, ok := m.timers[node.ID]; ok {
		timer.Stop()
	}
}

func (m *Manager) UpdateNodeWithExpire(node *NodeInfo, expire time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// do update
	m.doUpdateNode(node)

	// reset timer if it already exists
	nodeID := node.ID
	if timer, ok := m.timers[nodeID]; ok {
		timer.Reset(expire)
		return
	}

	// schedule a timer to remove this node after expiration
	m.timers[nodeID] = time.AfterFunc(expire, func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		if !m.nodes[nodeID].Alive {
			return
		}
		m.nodes[nodeID].Alive = false
		m.consistent.RemoveNode(nodeID)
		log.Printf("Node %v has been removed from consistent hash ring", node.ID)
	})
}

func (m *Manager) ExportHistory() map[uint64]*pb.NodeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.nodes.ToDTO()
}

func (m *Manager) ImportHistory(historyDTO map[uint64]*pb.NodeInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	history := NodeHistoryFromDTO(historyDTO)

	for nodeID, info := range history {
		curr, ok := m.nodes[nodeID]
		// update local node info if
		// 1. local info not exists
		// 2. local info has older version
		// 3. same version, but coming info suggests node not alive
		if !ok || curr.Version < info.Version || (curr.Version == info.Version && !info.Alive) {
			m.doUpdateNode(info)
		}
	}
}

func (m *Manager) GetPreferenceList(key []byte, num int) []uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.consistent.GetNodes(key, num)
}

func (m *Manager) GetInternalClient(nodeID uint64) (pb.KeyValueStoreInternalClient, error) {
	m.mu.RLock()

	// Get node info
	nodeInfo, ok := m.nodes[nodeID]
	if !ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("NodeManager: nodeID {%v} not found", nodeID)
	}

	// Reuse the existing connection
	if client, ok := m.internalPool[nodeInfo.InternalAddress]; ok {
		m.mu.RUnlock()
		return client, nil
	}

	// Create a new connection
	m.mu.RUnlock()
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.internalPool[nodeInfo.InternalAddress]; ok {
		return client, nil
	}

	conn, err := grpc.Dial(nodeInfo.InternalAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewKeyValueStoreInternalClient(conn)
	m.internalPool[nodeInfo.InternalAddress] = client
	return client, nil
}
