//go:build !(octopus || pacific || quincy || reef || squid)

package nvmegw

import (
	"github.com/ceph/go-ceph/internal/commands"
)

// GatewayList describes the status and details of the gateways for a pool and
// ANA-group.
type GatewayList struct {
	Epoch    int    `json:"epoch"`
	Pool     string `json:"pool"`
	Group    string `json:"group"`
	Features string `json:"features"`

	// below attributes are only set when a gateway is created

	RebalanceANAGroup int    `json:"rebalance_ana_group,omitempty"`
	GatewayEpoch      int    `json:"GW-epoch,omitempty"`
	ANAGroupList      string `json:"Anagrp list,omitempty"`

	Gateways []GatewayInfo `json:"Created Gateways:,omitempty"`
}

// GatewayInfo describes an NVMe-oF gateway.
type GatewayInfo struct {
	ID         string `json:"gw-id"`
	ANAGroupID int    `json:"anagrp-id"`

	PerformedFullStartup int    `json:"performed-full-startup"`
	Availability         string `json:"Availability"`
	ANAStates            string `json:"ana states"`

	// TODO: add structs for listeners and namespaces. This requires a
	// running and functional gateway. That is a tricky setup for the
	// testing environment.
}

func parseGatewayList(res commands.Response) (*GatewayList, error) {
	l := GatewayList{}
	if err := res.NoStatus().Unmarshal(&l).End(); err != nil {
		return nil, err
	}
	return &l, nil
}

// CreateGateway will create a new NVMe-oF gateway for accessing RBD-image.
//
// Similar To:
//
//	ceph nvme-gw create gw pool group
func (nvmea *Admin) CreateGateway(id, pool, group string) error {
	m := map[string]string{
		"prefix": "nvme-gw create",
		"format": "json",
		"id":     id,
		"pool":   pool,
		"group":  group,
	}

	return commands.MarshalMonCommand(nvmea.conn, m).NoData().End()
}

// DeleteGateway will delete a NVMe-oF gateway for accessing RBD-image.
//
// Similar To:
//
//	ceph nvme-gw delete gw pool group
func (nvmea *Admin) DeleteGateway(id, pool, group string) error {
	m := map[string]string{
		"prefix": "nvme-gw delete",
		"format": "json",
		"id":     id,
		"pool":   pool,
		"group":  group,
	}

	return commands.MarshalMonCommand(nvmea.conn, m).NoData().End()
}

// ShowGateways returns a list of gateways for the pool and group.
//
// Similar To:
//
//	ceph nvme-gw show pool group
func (nvmea *Admin) ShowGateways(pool, group string) (*GatewayList, error) {
	m := map[string]string{
		"prefix": "nvme-gw show",
		"format": "json",
		"pool":   pool,
		"group":  group,
	}
	return parseGatewayList(commands.MarshalMonCommand(nvmea.conn, m))
}
