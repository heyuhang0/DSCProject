package consistent

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)


func checkIntegrity(t *testing.T, c *Consistent, nodes []uint64) {
	// check number of virtual nodes
	assert.Len(t, c.ring, c.numVNodes*len(nodes),
		"wrong number of virtual nodes")

	// ring should be sorted
	assert.True(t, sort.IsSorted(c.ring),
		"virtual nodes are not sorted: %v", c.ring)

	for _, vNodeID := range c.ring {
		// check vNode -> node mapping
		nodeID := c.vNodeToNode[vNodeID]
		assert.Contains(t, nodes, nodeID)
		// check node -> vNodes mapping
		assert.Contains(t, c.nodeToVNodes[nodeID], vNodeID)
	}
}

func TestConsistent_AddRemoveNode(t *testing.T) {
	consistent := NewConsistent(5)

	// add nodes
	nodes := make([]uint64, 0)
	checkIntegrity(t, consistent, nodes)
	for nodeID := uint64(0); nodeID < 3; nodeID++ {
		consistent.AddNode(nodeID)
		nodes = append(nodes, nodeID)
		checkIntegrity(t, consistent, nodes)
	}

	// do nothing with duplicated node ID
	consistent.AddNode(nodes[0])
	checkIntegrity(t, consistent, nodes)

	// remove node
	toBeRemoved, nodes := nodes[0], nodes[1:]
	consistent.RemoveNode(toBeRemoved)
	checkIntegrity(t, consistent, nodes)

	toBeRemoved, nodes = nodes[0], nodes[1:]
	consistent.RemoveNode(toBeRemoved)
	checkIntegrity(t, consistent, nodes)
}

func TestConsistent_GetNodes(t *testing.T) {
	consistent := NewConsistent(5)

	// get nodes on an empty consistent hashing, expect empty result
	result := consistent.GetNodes("empty", 3)
	assert.Len(t, result, 0)

	// get nodes on ring with only one node, expect only that node
	consistent.AddNode(uint64(1))
	result = consistent.GetNodes("1node", 3)
	assert.ElementsMatch(t, []uint64{1}, result)

	// get 3 nodes on ring with 3 nodes, expect all nodes
	consistent.AddNode(uint64(2))
	consistent.AddNode(uint64(3))
	result = consistent.GetNodes("3node", 3)
	assert.ElementsMatch(t, []uint64{1, 2, 3}, result)

	// try different data type
	result = consistent.GetNodes(int(1), 3)
	assert.ElementsMatch(t, []uint64{1, 2, 3}, result)

	result = consistent.GetNodes(uint(1), 3)
	assert.ElementsMatch(t, []uint64{1, 2, 3}, result)

	result = consistent.GetNodes(uint32(2), 3)
	assert.ElementsMatch(t, []uint64{1, 2, 3}, result)

	result = consistent.GetNodes(uint64(2), 3)
	assert.ElementsMatch(t, []uint64{1, 2, 3}, result)

	// save some results
	result1 := consistent.GetNodes(1, 1)
	result2 := consistent.GetNodes(2, 2)
	result3 := consistent.GetNodes(3, 3)

	// result should be consistent
	assert.Equal(t, result1, consistent.GetNodes(1, 1))
	assert.Equal(t, result2, consistent.GetNodes(2, 2))
	assert.Equal(t, result3, consistent.GetNodes(3, 3))

	// try with more nodes
	consistent.AddNode(uint64(4))
	consistent.AddNode(uint64(5))
	consistent.GetNodes(1, 1)
	consistent.GetNodes(2, 2)
	consistent.GetNodes(3, 3)

	// try remove node
	consistent.RemoveNode(uint64(5))
	consistent.RemoveNode(uint64(4))
	// should give same result as before adding
	assert.Equal(t, result1, consistent.GetNodes(1, 1))
	assert.Equal(t, result2, consistent.GetNodes(2, 2))
	assert.Equal(t, result3, consistent.GetNodes(3, 3))
}

func BenchmarkConsistent_GetNodes(b *testing.B) {
	consistent := NewConsistent(10)
	for i := 0; i < 10; i ++ {
		consistent.AddNode(uint64(i))
	}
	for i := 0; i < b.N; i++ {
		consistent.GetNodes(i, 3)
	}
}

func TestConsistent_Redistribute(t *testing.T) {
	consistent := NewConsistent(10)
	for i := 0; i < 5; i ++ {
		consistent.AddNode(uint64(i))
	}

	// helpers
	getMapping := func() map[int]uint64 {
		mapping := make(map[int]uint64)
		for key := 0; key < 10000; key ++ {
			mapping[key] = consistent.GetNode(key)
		}
		return mapping
	}
	countLoad := func(mapping map[int]uint64) map[uint64]int {
		load := make(map[uint64]int)
		for _, node := range mapping {
			if _, ok := load[node]; ok {
				load[node] += 1
				continue
			}
			load[node] = 1
		}
		return load
	}
	countRedistributed := func(mappingBefore, mappingAfter map[int]uint64) map[uint64]int {
		redistributed := make(map[uint64]int)
		for k, nodeBefore := range mappingBefore {
			nodeAfter := mappingAfter[k]
			if nodeBefore != nodeAfter {
				if _, ok := redistributed[nodeBefore]; ok {
					redistributed[nodeBefore] += 1
					continue
				}
				redistributed[nodeBefore] = 1
			}
		}
		return redistributed
	}

	// 5 nodes
	mapping5 := getMapping()
	fmt.Println("With 5 nodes:")
	fmt.Println("Nodes Load\t\t", countLoad(mapping5))

	// add one node (6 nodes)
	consistent.AddNode(uint64(5))
	mapping6 := getMapping()
	fmt.Println("With 6 nodes:")
	fmt.Println("Nodes Load\t\t", countLoad(mapping6))
	fmt.Println("Redistributed\t", countRedistributed(mapping5, mapping6))

	// add one node (7 nodes)
	consistent.AddNode(uint64(6))
	mapping7 := getMapping()
	fmt.Println("With 7 nodes:")
	fmt.Println("Nodes Load\t\t", countLoad(mapping7))
	fmt.Println("Redistributed\t", countRedistributed(mapping6, mapping7))
}

func TestConsistent_ToDTO(t *testing.T) {
	consistent := NewConsistent(10)
	for i := 0; i < 5; i ++ {
		consistent.AddNode(uint64(i))
	}
	consistentFromDTO := FromDTO(consistent.ToDTO())

	for i := 0; i < 1000; i ++ {
		assert.Equal(t, consistent.GetNodes(i, 3), consistentFromDTO.GetNodes(i, 3))
	}
}
