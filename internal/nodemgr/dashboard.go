package nodemgr

import (
	"encoding/json"
	"fmt"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"log"
	"net/http"
)

type dashboardNodeInfo struct {
	*pb.NodeInfo
	VirtualNodes []uint64 `json:"virtualNodes"`
}

type dashboardState struct {
	Nodes map[uint64]*dashboardNodeInfo `json:"nodes"`
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

		virtualNodesOfNode := m.consistent.ExportVirtualNodes()
		for key, nodeInfo := range history {
			var virtualNodes []uint64
			var ok bool
			if virtualNodes, ok = virtualNodesOfNode[key]; !ok {
				virtualNodes = make([]uint64, 0)
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
