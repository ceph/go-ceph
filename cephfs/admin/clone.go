// +build !luminous,!mimic

package admin

// CloneOptions are used to specify optional values to be used when creating a
// new subvolume clone.
type CloneOptions struct {
	TargetGroup string
	PoolLayout  string
}

// CloneSubVolumeSnapshot Clones the specified snapshot from the subvolume.
//
// Similar To:
//  ceph fs subvolume snapshot clone <volume> <subvolume> <snapshot> <name>
func (fsa *FSAdmin) CloneSubVolumeSnapshot(volume, group, subvolume, snapshot, name string, o *CloneOptions) error {
	m := map[string]string{
		"prefix":          "fs subvolume snapshot clone",
		"vol_name":        volume,
		"sub_name":        subvolume,
		"snap_name":       snapshot,
		"target_sub_name": name,
		"format":          "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	if o != nil && o.TargetGroup != NoGroup {
		m["target_group_name"] = group
	}
	if o != nil && o.PoolLayout != "" {
		m["pool_layout"] = o.PoolLayout
	}
	return checkEmptyResponseExpected(fsa.marshalMgrCommand(m))
}

// CloneState is used to define constant values used to determine the state of
// a clone.
type CloneState string

const (
	// ClonePending is the state of a pending clone.
	ClonePending = CloneState("pending")
	// CloneInProgress is the state of a clone in progress.
	CloneInProgress = CloneState("in-progress")
	// CloneComplete is the state of a complete clone.
	CloneComplete = CloneState("complete")
	// CloneFailed is the state of a failed clone.
	CloneFailed = CloneState("failed")
)

// CloneSource contains values indicating the source of a clone.
type CloneSource struct {
	Volume    string `json:"volume"`
	Group     string `json:"group"`
	SubVolume string `json:"subvolume"`
	Snapshot  string `json:"snapshot"`
}

// CloneStatus reports on the status of a subvolume clone.
type CloneStatus struct {
	State  CloneState  `json:"state"`
	Source CloneSource `json:"source"`
}

type cloneStatusWrapper struct {
	Status CloneStatus `json:"status"`
}

func parseCloneStatus(r []byte, s string, err error) (*CloneStatus, error) {
	var status cloneStatusWrapper
	if err := unmarshalResponseJSON(r, s, err, &status); err != nil {
		return nil, err
	}
	return &status.Status, nil
}

// CloneStatus returns data reporting the status of a subvolume clone.
//
// Similar To:
//  ceph fs clone status <volume> --group_name=<group> <clone>
func (fsa *FSAdmin) CloneStatus(volume, group, clone string) (*CloneStatus, error) {
	m := map[string]string{
		"prefix":     "fs clone status",
		"vol_name":   volume,
		"clone_name": clone,
		"format":     "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return parseCloneStatus(fsa.marshalMgrCommand(m))
}

// CancelClone stops the background processes that populate a clone.
// CancelClone does not delete the clone.
//
// Similar To:
//  ceph fs clone cancel <volume> --group_name=<group> <clone>
func (fsa *FSAdmin) CancelClone(volume, group, clone string) error {
	m := map[string]string{
		"prefix":     "fs clone cancel",
		"vol_name":   volume,
		"clone_name": clone,
		"format":     "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return checkEmptyResponseExpected(fsa.marshalMgrCommand(m))
}
