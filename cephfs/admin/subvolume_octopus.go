// +build octopus

package admin

// SubVolumeSnapshotInfo reports various informational values about a subvolume.
type SubVolumeSnapshotInfo struct {
	CreatedAt        TimeStamp `json:"created_at"`
	DataPool         string    `json:"data_pool"`
	HasPendingClones string    `json:"has_pending_clones"`
	Protected        string    `json:"protected"`
	Size             ByteCount `json:"size"`
}

func parseSubVolumeSnapshotInfo(r []byte, s string, err error) (*SubVolumeSnapshotInfo, error) {
	var info SubVolumeSnapshotInfo
	if err := unmarshalResponseJSON(r, s, err, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// SubVolumeSnapshotInfo returns information about the specified subvolume snapshot.
//
// Similar To:
//  ceph fs subvolume snapshot info <volume> --group-name=<group> <subvolume> <name>
func (fsa *FSAdmin) SubVolumeSnapshotInfo(volume, group, subvolume, name string) (*SubVolumeSnapshotInfo, error) {
	m := map[string]string{
		"prefix":    "fs subvolume snapshot info",
		"vol_name":  volume,
		"sub_name":  subvolume,
		"snap_name": name,
		"format":    "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return parseSubVolumeSnapshotInfo(fsa.marshalMgrCommand(m))
}
