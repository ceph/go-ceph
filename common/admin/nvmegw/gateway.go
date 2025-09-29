//go:build !(octopus || pacific || quincy || reef) && ceph_preview

package nvmegw

import (
	"github.com/ceph/go-ceph/internal/commands"
)

type GatewayList struct {
	Epoch    int    `json:"epoch"`
	Pool     string `json:"pool"`
	Group    string `json:"group"`
	Features string `json:"features"`

	// below attributes are only set when a gateway is created

	// RebalanceANAGroup ...
	RebalanceANAGroup int `json:"rebalance_ana_group,omitempty"`
	GatewayEpoch      int `json:"GW-epoch,omitempty"`
	// ANAGroupList is in the format of "[ 1 ]"
	ANAGroupList string `json:"Anagrp list,omitempty"`

	// there is also "num-namespaces": 0,

	// Gateways is missing when "num gws": 0
	Gateways []GatewayInfo `json:"Created Gateways,omitempty"`
}

// GatewayInfo describes an NVMe-oF gateway.
type GatewayInfo struct {
	ID         string `json:"gw-id"`
	ANAGroupID int    `json:"anagrp-id"`

	// "num-namespaces": 0
	PerformedFullStartup bool `json:"performed-full-startup"`
	// "CREATED",
	Availability string `json:"Availability"`
	// formatted like " 1: STANDBY "
	ANAStates string `json:"ana states"`

	// num-listeners (int) is only set when listeners are configured?
}

func parseGatewayList(res commands.Response) ([]GatewayInfo, error) {
	l := []GatewayInfo{}
	if err := res.NoStatus().Unmarshal(&l).End(); err != nil {
		return nil, err
	}
	return l, nil
}

// CreateNVMeGateway will create a new NVMe-oF gateway for accessing RBD-image.
//
// Similar To:
//
//	ceph nvme-gw create gw pool group
func (nvmea *Admin) CreateNVMeGateway(id, pool, group string) error {
	m := map[string]string{
		"prefix": "nvme-gw create",
		"format": "json",
		"id":     id,
		"pool":   pool,
		"group":  group,
	}

	return commands.MarshalMgrCommand(nvmea.conn, m).NoData().End()
}

// DeleteNVMeGateway will delete a NVMe-oF gateway for accessing RBD-image.
//
// Similar To:
//
//	ceph nvme-gw delete gw pool group
func (nvmea *Admin) DeleteNVMeGateway(id, pool, group string) error {
	m := map[string]string{
		"prefix": "nvme-gw delete",
		"format": "json",
		"id":     id,
		"pool":   pool,
		"group":  group,
	}

	return commands.MarshalMgrCommand(nvmea.conn, m).NoData().End()
}

// ShowGateways returns a list of gateways for the pool and group.
//
// Similar To:
//
//	ceph nvme-gw show pool group
func (nvmea *Admin) ShowGateways(pool, group string) ([]GatewayInfo, error) {
	m := map[string]string{
		"prefix": "nvme-gw show",
		"format": "json",
		"pool":   pool,
		"group":  group,
	}
	return parseGatewayList(commands.MarshalMgrCommand(nvmea.conn, m))
}
