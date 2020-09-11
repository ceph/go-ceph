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

func parseVolumeStatus(res []byte, status string, err error) (*VolumeStatus, error) {
	var vs VolumeStatus
	err = unmarshalResponseJSON(res, status, err, &vs)
	return &vs, err
}

// VolumeStatus returns a VolumeStatus object for the given volume name.
//
// Similar To:
//  ceph fs status cephfs <name>
func (fsa *FSAdmin) VolumeStatus(name string) (*VolumeStatus, error) {
	r, s, err := fsa.marshalMgrCommand(map[string]string{
		"fs":     name,
		"prefix": "fs status",
		"format": "json",
	})
	return parseVolumeStatus(r, s, err)
}
