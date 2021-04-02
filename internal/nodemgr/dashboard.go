package nodemgr

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/cespare/xxhash"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"log"
	"net/http"
)

type dashboardNodeInfo struct {
	*pb.NodeInfo
	VirtualNodes []uint64 `json:"virtualNodes,omitempty"`
}

type dashboardState struct {
	Nodes map[uint64]*dashboardNodeInfo `json:"nodes,omitempty"`
}

func hashUint64(val uint64) uint64 {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, val)
	return xxhash.Sum64(bs)
}

func (m *Manager) ServeDashboard(addr string) error {
	// serve static files
	fs := http.FileServer(http.Dir("./web/visualization"))
	http.Handle("/", fs)

	// server api
	http.HandleFunc("/api/dashboard", func(writer http.ResponseWriter, request *http.Request) {
		history := m.ExportHistory()

		state := &dashboardState{
			Nodes: make(map[uint64]*dashboardNodeInfo),
		}

		for key, nodeInfo := range history {
			virtualNodes := make([]uint64, m.numVNodes)
			virtualNodes[0] = hashUint64(nodeInfo.Id)
			for i := 1; i < len(virtualNodes); i ++ {
				virtualNodes[i] = hashUint64(virtualNodes[i-1])
			}
			state.Nodes[key] = &dashboardNodeInfo{
				NodeInfo:     nodeInfo,
				VirtualNodes: virtualNodes,
			}
		}

		stateJson, err := json.Marshal(state)
		if err != nil {
			log.Print(err)
			writer.WriteHeader(500)
			return
		}
		writer.Header().Add("Content-Type", "application/json")
		_, _ = fmt.Fprint(writer, string(stateJson))
	})

	return http.ListenAndServe(addr, nil)
}
