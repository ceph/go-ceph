package admin

import (
	"encoding/json"
	"fmt"
)

var listVolumesCmd = []byte(`{"prefix":"fs volume ls"}`)

// ListVolumes return a list of volumes in this Ceph cluster.
func (fsa *FSAdmin) ListVolumes() ([]string, error) {
	r, s, err := fsa.rawMgrCommand(listVolumesCmd)
	return parseListNames(r, s, err)
}

// VolumeStatus reports various properties of a CephFS volume.
// TODO: Fill in.
type VolumeStatus struct {
	MDSVersion string `json:"mds_version"`
}

func parseVolumeStatus(res []byte, status string, err error) (*VolumeStatus, error) {
	if err != nil {
		return nil, err
	}
	if status != "" {
		return nil, fmt.Errorf("error status: %s", status)
	}
	var vs VolumeStatus
	if err := json.Unmarshal(res, &vs); err != nil {
		return nil, err
	}
	return &vs, nil
}

// VolumeStatus returns a VolumeStatus object for the given volume name.
func (fsa *FSAdmin) VolumeStatus(name string) (*VolumeStatus, error) {
	r, s, err := fsa.marshalMgrCommand(map[string]string{
		"fs":     name,
		"prefix": "fs status",
		"format": "json",
	})
	return parseVolumeStatus(r, s, err)
}
