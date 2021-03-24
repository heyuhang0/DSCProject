package consistent

import (
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
