// +build octopus

package admin

// VolumePool reports on the pool status for a CephFS volume.
type VolumePool struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Available uint64 `json:"avail"`
	Used      uint64 `json:"used"`
}

// VolumeStatus reports various properties of a CephFS volume.
// TODO: Fill in.
type VolumeStatus struct {
	MDSVersion string       `json:"mds_version"`
	Pools      []VolumePool `json:"pools"`
}

func parseVolumeStatus(res response) (*VolumeStatus, error) {
	var vs VolumeStatus
	err := res.noStatus().unmarshal(&vs).End()
	return &vs, err
}

// VolumeStatus returns a VolumeStatus object for the given volume name.
//
// Similar To:
//  ceph fs status cephfs <name>
func (fsa *FSAdmin) VolumeStatus(name string) (*VolumeStatus, error) {
	res := fsa.marshalMgrCommand(map[string]string{
		"fs":     name,
		"prefix": "fs status",
		"format": "json",
	})
	return parseVolumeStatus(res)
}
