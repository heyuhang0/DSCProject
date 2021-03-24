package consistent

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

type Hashing interface {
	AddNode(nodeID uint64)
	RemoveNode(nodeID uint64)
	GetNode(key interface{}) uint64
	GetNodes(key interface{}, num int) []uint64
}

type uint64Slice []uint64

type Consistent struct {
	mu sync.RWMutex

	ring         uint64Slice
	numVNodes    int
	vNodeToNode  map[uint64]uint64
	nodeToVNodes map[uint64][]uint64
	hash         func(key []byte) uint64
}

func (r uint64Slice) Len() int           { return len(r) }
func (r uint64Slice) Less(i, j int) bool { return r[i] < r[j] }
func (r uint64Slice) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func NewConsistent(numVNodes int) *Consistent {
	hash := func(key []byte) uint64 {
		h := fnv.New64a()
		_, _ = h.Write(key)
		return h.Sum64()
	}
	return NewConsistentWithHash(numVNodes, hash)
}

func NewConsistentWithHash(numVNodes int, hash func(key []byte) uint64) *Consistent {
	return &Consistent{
		ring:         make(uint64Slice, 0),
		numVNodes:    numVNodes,
		vNodeToNode:  make(map[uint64]uint64),
		nodeToVNodes: make(map[uint64][]uint64),
		hash:         hash,
	}
}

func insert(slice []uint64, index int, value uint64) []uint64 {
	if len(slice) == index {
		return append(slice, value)
	}
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = value
	return slice
}

func (r *Consistent) hashUInt64(key uint64) uint64 {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, key)
	return r.hash(bs)
}

func (r *Consistent) hashSupported(key interface{}) uint64 {
	switch v := key.(type) {
	case int:
		return r.hashUInt64(uint64(v))
	case uint:
		return r.hashUInt64(uint64(v))
	case string:
		return r.hash([]byte(v))
	}

	// binary.Write() can handle a fixed-size value,
	// or a slice of fixed-size values,
	// or a pointer to such data.
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, key)
	if err != nil {
		// binary.Write to a Buffer should never fail
		// with supported type, so error should not
		// be considered as a runtime error
		panic(fmt.Sprint("binary.Write failed: ", err))
	}
	return r.hash(buf.Bytes())
}

// find the index of the first VNode with ID >= key, len(r.ring) if no such node
func (r *Consistent) findVNode(key uint64) int {
	return sort.Search(len(r.ring), func(i int) bool { return r.ring[i] >= key })
}

func (r *Consistent) addVNode(vNodeID uint64) {
	index := r.findVNode(vNodeID)
	r.ring = insert(r.ring, index, vNodeID)
}

func (r *Consistent) AddNode(nodeID uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// do nothing if node already exists
	if _, exists := r.nodeToVNodes[nodeID]; exists {
		return
	}

	// generate a list of virtual nodes
	vNodes := make([]uint64, r.numVNodes)
	vNodeID := nodeID
	for i := range vNodes {
		vNodeID = r.hashUInt64(vNodeID)
		if _, exists := r.vNodeToNode[vNodeID]; exists {
			// This Case Should NOT Happen in Practice
			//
			// Since we are using 64-bit hash (assume uniformed),
			// even with 6100 virtual nodes,
			// the probability of collision is < 10^(-12).
			// see Birthday Attack for more details
			//
			// And the consistent hashing is only controlled by the system,
			// so it is exposed to external hash collision attack.
			//
			// Due to the low possibility, just let it CRASH.
			// Because handling it may cause more problems,
			// for example, if rehashing is used, adding nodes in different
			// order may result in inconsistent hash function
			panic("duplicated vNode")
		}
		vNodes[i] = vNodeID

		// save mapping from virtual node to node
		r.vNodeToNode[vNodeID] = nodeID

		// add virtual node to Consistent
		r.addVNode(vNodeID)
	}

	// save mapping from node to virtual nodes
	r.nodeToVNodes[nodeID] = vNodes
}

func remove(slice []uint64, index int) []uint64 {
	return append(slice[:index], slice[index+1:]...)
}

func (r *Consistent) removeVNode(vNodeID uint64) {
	index := r.findVNode(vNodeID)
	r.ring = remove(r.ring, index)
}

func (r *Consistent) RemoveNode(nodeID uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	vNodes, ok := r.nodeToVNodes[nodeID]
	if !ok {
		return
	}

	for _, vNodeID := range vNodes {
		// remove node from ring
		r.removeVNode(vNodeID)

		// delete vNode to Node mapping
		delete(r.vNodeToNode, vNodeID)
	}

	// delete Node to VNodes mapping
	delete(r.nodeToVNodes, nodeID)
}

func (r *Consistent) GetNode(key interface{}) uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := r.GetNodes(key, 1)
	if len(nodes) != 1 {
		panic("getting node on an empty consistent hashing")
	}
	return nodes[0]
}

func (r *Consistent) GetNodes(key interface{}, num int) []uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// return empty list if empty
	numNodes := len(r.nodeToVNodes)
	if numNodes == 0 {
		return make([]uint64, 0)
	}
	if numNodes < num {
		num = numNodes
	}

	// get key hash in uint64
	keyHash := r.hashSupported(key)

	// find the first node
	numVNodes := len(r.vNodeToNode)
	firstVNodeIndex := r.findVNode(keyHash) % numVNodes

	// find list of nodes
	nodes := make([]uint64, 0, num)
	nodeSet := make(map[uint64]interface{})
	for i := 0; len(nodes) < num; i++ {
		vNodeIndex := (firstVNodeIndex + i) % numVNodes
		node := r.vNodeToNode[r.ring[vNodeIndex]]

		// prevent duplicate physical node
		if _, exists := nodeSet[node]; exists {
			continue
		}

		nodes = append(nodes, node)
		nodeSet[node] = nil
	}
	return nodes
}
