//go:build ceph_preview
// +build ceph_preview

package telemetry

import (
	"fmt"

	ccom "github.com/ceph/go-ceph/common/commands"
	"github.com/ceph/go-ceph/internal/commands"
)

// Admin is used to administer ceph nfs features.
type Admin struct {
	conn ccom.RadosCommander
}

// NewFromConn creates an new management object from a preexisting
// rados connection. The existing connection can be rados.Conn or any
// type implementing the RadosCommander interface.
func NewFromConn(conn ccom.RadosCommander) *Admin {
	return &Admin{conn}
}

// Status is the output of "ceph telemetry status"
type Status struct {
	ChannelBasic      bool   `json:"Channel_basic"`
	ChannelCrash      bool   `json:"Channel_crash"`
	ChannelDevice     bool   `json:"Channel_device"`
	ChannelIdent      bool   `json:"Channel_ident"`
	ChannelPerf       bool   `json:"Channel_perf"`
	Contact           string `json:"Contact"`
	Description       string `json:"Description"`
	DeviceURL         string `json:"Device_url"`
	Enabled           bool   `json:"Enabled"`
	Interval          int    `json:"Interval"`
	LastOptRevision   int    `json:"Last_opt_revision"`
	LastUpload        string `json:"Last_upload"`
	Leaderboard       bool   `json:"Leaderboard"`
	LogLevel          string `json:"Log_level"`
	LogToCluster      bool   `json:"Log_to_cluster"`
	LogToClusterLevel string `json:"Log_to_cluster_level"`
	LogToFile         bool   `json:"Log_to_file"`
	Organization      string `json:"Organization"`
	Proxy             string `json:"Proxy"`
	URL               string `json:"Url"`
}

func parseStatus(res commands.Response) (*Status, error) {
	m := &Status{}
	if err := res.NoStatus().Unmarshal(m).End(); err != nil {
		return nil, err
	}
	return m, nil
}

// Status returns the output of "ceph telemetry status"
func (fsa *Admin) Status() (*Status, error) {
	m := map[string]string{
		"prefix": "telemetry status",
		"format": "json",
	}
	return parseStatus(commands.MarshalMonCommand(fsa.conn, m))
}

// Data returns the output of the telemetry module
// For this is either uses "preview-all" or "show-all"
//
//	depending if the module if activated or not
func (fsa *Admin) Data() ([]byte, error) {
	status, err := fsa.Status()
	if err != nil {
		return nil, err
	}

	commandArg := "preview"
	if status.Enabled {
		commandArg = "show"
	}

	m := map[string]string{
		"prefix": fmt.Sprintf("telemetry %s-all", commandArg),
		"format": "json",
	}
	resp := commands.MarshalMonCommand(fsa.conn, m)
	if resp.Unwrap() != nil {
		return nil, resp.Unwrap()
	}
	return resp.Body(), nil
}
