package api

import (
	"fmt"
)

type CephNode struct {
	Children        []int   `json:"children"`
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	TypeID          int     `json:"type_id"`
	CrushWeight     float64 `json:"crush_weight"`
	Depth           int     `json:"depth"`
	Exists          int     `json:"exists"`
	PrimaryAffinity float64 `json:"primary_affinity"`
	Reweight        float64 `json:"reweight"`
	Status          string  `json:"status"`
}

type OsdTree struct {
	Output struct {
		Nodes []CephNode    `json:"nodes"`
		Stray []interface{} `json:"stray"`
	} `json:"output"`
	Status string `json:"status"`
}

// Gets a map of id to cephNode, which can be used to index
func (ot *OsdTree) GetCephNodeMapById() map[int]CephNode {
	nodeMap := make(map[int]CephNode)
	for _, node := range ot.Output.Nodes {
		nodeMap[node.ID] = node
	}
	return nodeMap
}

// Gets a map of id to cephNode, which can be used to index
func (ot *OsdTree) GetCephNodeMapByName() map[string]CephNode {
	nodeMap := make(map[string]CephNode)
	for _, node := range ot.Output.Nodes {
		nodeMap[node.Name] = node
	}
	return nodeMap
}

func (ot *OsdTree) FilterNodes(filter string) []CephNode {
	var cephNodes []CephNode
	for _, node := range ot.Output.Nodes {
		if node.Type == filter {
			cephNodes = append(cephNodes, node)
		}
	}
	return cephNodes
}

func getActiveState(reweight float64) string {
	if reweight == 0.0 {
		return "out"
	}
	return "in"
}

func filterOsdsByState(cephNodes []CephNode, liveState string, activeState string) []CephNode {
	var nodes []CephNode
	for _, node := range cephNodes {
		if node.Status == liveState && getActiveState(node.Reweight) == activeState {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (ot *OsdTree) GetOsdsByState(liveState, activeState string) []CephNode {
	return filterOsdsByState(ot.FilterNodes("osd"), liveState, activeState)
}

func (ot *OsdTree) GetOsdsByHost(hostname, liveState, activeState string) ([]CephNode, error) {
	nameNodeMap := ot.GetCephNodeMapByName()
	idNodeMap := ot.GetCephNodeMapById()

	node, exists := nameNodeMap[hostname]
	var nodes []CephNode
	if !exists {
		return nodes, fmt.Errorf("Could not find ceph hostname %s", hostname)
	}
	for _, id := range node.Children {
		nodes = append(nodes, idNodeMap[id])
	}
	return filterOsdsByState(nodes, liveState, activeState), nil
}

func (ot *OsdTree) GetCephNodes() []CephNode {
	return ot.Output.Nodes
}

func (ot *OsdTree) GetCephNodeByName(name string) (CephNode, error) {
	nodeMap := ot.GetCephNodeMapByName()
	node, exists := nodeMap[name]
	if !exists {
		return node, fmt.Errorf("Could not find ceph node with name %s", name)
	}
	return node, nil
}

func (ot *OsdTree) GetCephNodeById(id int) (CephNode, error) {
	nodeMap := ot.GetCephNodeMapById()
	node, exists := nodeMap[id]
	if !exists {
		return node, fmt.Errorf("Could not find ceph node with id %d", id)
	}
	return node, nil
}
