package api

import (
	"encoding/json"
	"fmt"
)

func (cc *CephClient) GetStatus() (Status, error) {
	var status Status
	body, err := cc.callApi("status", "GET")
	if err != nil {
		return status, err
	}
	err = json.Unmarshal([]byte(body), &status)
	if err != nil {
		return status, err
	}
	return status, nil
}

func (cc *CephClient) GetOsdTree() (OsdTree, error) {
	var osdTree OsdTree
	body, err := cc.callApi("osd/tree", "GET")
	if err != nil {
		return osdTree, err
	}
	err = json.Unmarshal([]byte(body), &osdTree)
	if err != nil {
		return osdTree, err
	}
	return osdTree, nil
}

// Filter by type
func (cc *CephClient) GetCephNodes(filter string) ([]CephNode, error) {
	osdTree, err := cc.GetOsdTree();
	if err != nil {
		return []CephNode{}, err
	}
	if filter == "" {
		return osdTree.Output.Nodes, nil
	}
	var finalNodes [] CephNode 
	for _, node := range osdTree.Output.Nodes {
		if node.Type == filter {
			finalNodes = append(finalNodes, node)
		}
	}
	return finalNodes, nil
}

func (cc *CephClient) GetCephNodeByName(name string) (CephNode, error) {
	allNodes, err := cc.GetCephNodes("")
	if err != nil {
		return CephNode{}, err 
	}

	for _, node := range allNodes {
		if node.Name == name {
			return node, nil
		}
	}
	return CephNode{}, fmt.Errorf("Node with name %s does not exist", name)
}

func (cc *CephClient) GetCephNodeById(id int) (CephNode, error) {
	allNodes, err := cc.GetCephNodes("")
	if err != nil {
		return CephNode{}, err 
	}

	for _, node := range allNodes {
		if node.ID == id {
			return node, nil
		}
	}
	return CephNode{}, fmt.Errorf("Node with id %d does not exist", id)
}

func (cc *CephClient) GetBlacklist() (OsdBlacklistLs, error) {
	var osdBlacklistLs OsdBlacklistLs
	body, err := cc.callApi("osd/blacklist/ls", "GET")
	if err != nil {
		return osdBlacklistLs, err
	}
	err = json.Unmarshal([]byte(body), &osdBlacklistLs)
	if err != nil {
		return osdBlacklistLs, err
	}
	return osdBlacklistLs, nil
}

func (cc *CephClient) GetMdsStat() (MdsStat, error) {
	var mdsStatus MdsStat
	body, err := cc.callApi("mds/stat", "GET")
	if err != nil {
		return mdsStatus, err
	}
	fmt.Printf("Body: \n %s", body)
	err = json.Unmarshal([]byte(body), &mdsStatus)
	if err != nil {
		return mdsStatus, err
	}
	return mdsStatus, nil
}


