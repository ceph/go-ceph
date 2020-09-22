// +build !luminous,!mimic

package admin

// this is the internal type used to create JSON for ceph.
// See SubVolumeOptions for the type that users of the library
// interact with.
// note that the ceph json takes mode as a string.
type subVolumeFields struct {
	Prefix            string    `json:"prefix"`
	Format            string    `json:"format"`
	VolName           string    `json:"vol_name"`
	GroupName         string    `json:"group_name,omitempty"`
	SubName           string    `json:"sub_name"`
	Size              ByteCount `json:"size,omitempty"`
	Uid               int       `json:"uid,omitempty"`
	Gid               int       `json:"gid,omitempty"`
	Mode              string    `json:"mode,omitempty"`
	PoolLayout        string    `json:"pool_layout,omitempty"`
	NamespaceIsolated bool      `json:"namespace_isolated"`
}

// SubVolumeOptions are used to specify optional, non-identifying, values
// to be used when creating a new subvolume.
type SubVolumeOptions struct {
	Size              ByteCount
	Uid               int
	Gid               int
	Mode              int
	PoolLayout        string
	NamespaceIsolated bool
}

func (s *SubVolumeOptions) toFields(v, g, n string) *subVolumeFields {
	return &subVolumeFields{
		Prefix:            "fs subvolume create",
		Format:            "json",
		VolName:           v,
		GroupName:         g,
		SubName:           n,
		Size:              s.Size,
		Uid:               s.Uid,
		Gid:               s.Gid,
		Mode:              modeString(s.Mode, false),
		PoolLayout:        s.PoolLayout,
		NamespaceIsolated: s.NamespaceIsolated,
	}
}

// NoGroup should be used when an optional subvolume group name is not
// specified.
const NoGroup = ""

// CreateSubVolume sends a request to create a CephFS subvolume in a volume,
// belonging to an optional subvolume group.
//
// Similar To:
//  ceph fs subvolume create <volume> --group-name=<group> <name> ...
func (fsa *FSAdmin) CreateSubVolume(volume, group, name string, o *SubVolumeOptions) error {
	if o == nil {
		o = &SubVolumeOptions{}
	}
	f := o.toFields(volume, group, name)
	return fsa.marshalMgrCommand(f).noData().End()
}

// ListSubVolumes returns a list of subvolumes belonging to the volume and
// optional subvolume group.
//
// Similar To:
//  ceph fs subvolume ls <volume> --group-name=<group>
func (fsa *FSAdmin) ListSubVolumes(volume, group string) ([]string, error) {
	m := map[string]string{
		"prefix":   "fs subvolume ls",
		"vol_name": volume,
		"format":   "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return parseListNames(fsa.marshalMgrCommand(m))
}

// RemoveSubVolume will delete a CephFS subvolume in a volume and optional
// subvolume group.
//
// Similar To:
//  ceph fs subvolume rm <volume> --group-name=<group> <name> ...
func (fsa *FSAdmin) RemoveSubVolume(volume, group, name string) error {
	m := map[string]string{
		"prefix":   "fs subvolume rm",
		"vol_name": volume,
		"sub_name": name,
		"format":   "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return fsa.marshalMgrCommand(m).noData().End()
}

type subVolumeResizeFields struct {
	Prefix    string `json:"prefix"`
	Format    string `json:"format"`
	VolName   string `json:"vol_name"`
	GroupName string `json:"group_name,omitempty"`
	SubName   string `json:"sub_name"`
	NewSize   string `json:"new_size"`
	NoShrink  bool   `json:"no_shrink"`
}

// SubVolumeResizeResult reports the size values returned by the
// ResizeSubVolume function, as reported by Ceph.
type SubVolumeResizeResult struct {
	BytesUsed    ByteCount `json:"bytes_used"`
	BytesQuota   ByteCount `json:"bytes_quota"`
	BytesPercent string    `json:"bytes_pcent"`
}

// ResizeSubVolume will resize a CephFS subvolume. The newSize value may be a
// ByteCount or the special Infinite constant. Setting noShrink to true will
// prevent reducing the size of the volume below the current used size.
//
// Similar To:
//  ceph fs subvolume resize <volume> --group-name=<group> <name> ...
func (fsa *FSAdmin) ResizeSubVolume(
	volume, group, name string,
	newSize QuotaSize, noShrink bool) (*SubVolumeResizeResult, error) {

	f := &subVolumeResizeFields{
		Prefix:    "fs subvolume resize",
		Format:    "json",
		VolName:   volume,
		GroupName: group,
		SubName:   name,
		NewSize:   newSize.resizeValue(),
		NoShrink:  noShrink,
	}
	var result []*SubVolumeResizeResult
	res := fsa.marshalMgrCommand(f)
	if err := res.noStatus().unmarshal(&result).End(); err != nil {
		return nil, err
	}
	return result[0], nil
}

// SubVolumePath returns the path to the subvolume from the root of the file system.
//
// Similar To:
//  ceph fs subvolume getpath <volume> --group-name=<group> <name>
func (fsa *FSAdmin) SubVolumePath(volume, group, name string) (string, error) {
	m := map[string]string{
		"prefix":   "fs subvolume getpath",
		"vol_name": volume,
		"sub_name": name,
		// ceph doesn't respond in json for this cmd (even if you ask)
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return parsePathResponse(fsa.marshalMgrCommand(m))
}

// SubVolumeInfo reports various informational values about a subvolume.
type SubVolumeInfo struct {
	Type          string    `json:"type"`
	Path          string    `json:"path"`
	Uid           int       `json:"uid"`
	Gid           int       `json:"gid"`
	Mode          int       `json:"mode"`
	BytesPercent  string    `json:"bytes_pcent"`
	BytesUsed     ByteCount `json:"bytes_used"`
	BytesQuota    QuotaSize `json:"-"`
	DataPool      string    `json:"data_pool"`
	PoolNamespace string    `json:"pool_namespace"`
	Atime         TimeStamp `json:"atime"`
	Mtime         TimeStamp `json:"mtime"`
	Ctime         TimeStamp `json:"ctime"`
	CreatedAt     TimeStamp `json:"created_at"`
}

type subVolumeInfoWrapper struct {
	SubVolumeInfo
	VBytesQuota *quotaSizePlaceholder `json:"bytes_quota"`
}

func parseSubVolumeInfo(res response) (*SubVolumeInfo, error) {
	var info subVolumeInfoWrapper
	if err := res.noStatus().unmarshal(&info).End(); err != nil {
		return nil, err
	}
	if info.VBytesQuota != nil {
		info.BytesQuota = info.VBytesQuota.Value
	}
	return &info.SubVolumeInfo, nil
}

// SubVolumeInfo returns information about the specified subvolume.
//
// Similar To:
//  ceph fs subvolume info <volume> --group-name=<group> <name>
func (fsa *FSAdmin) SubVolumeInfo(volume, group, name string) (*SubVolumeInfo, error) {
	m := map[string]string{
		"prefix":   "fs subvolume info",
		"vol_name": volume,
		"sub_name": name,
		"format":   "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return parseSubVolumeInfo(fsa.marshalMgrCommand(m))
}

// CreateSubVolumeSnapshot creates a new snapshot from the source subvolume.
//
// Similar To:
//  ceph fs subvolume snapshot create <volume> --group-name=<group> <source> <name>
func (fsa *FSAdmin) CreateSubVolumeSnapshot(volume, group, source, name string) error {
	m := map[string]string{
		"prefix":    "fs subvolume snapshot create",
		"vol_name":  volume,
		"sub_name":  source,
		"snap_name": name,
		"format":    "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return fsa.marshalMgrCommand(m).noData().End()
}

// RemoveSubVolumeSnapshot removes the specified snapshot from the subvolume.
//
// Similar To:
//  ceph fs subvolume snapshot rm <volume> --group-name=<group> <subvolume> <name>
func (fsa *FSAdmin) RemoveSubVolumeSnapshot(volume, group, subvolume, name string) error {
	m := map[string]string{
		"prefix":    "fs subvolume snapshot rm",
		"vol_name":  volume,
		"sub_name":  subvolume,
		"snap_name": name,
		"format":    "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return fsa.marshalMgrCommand(m).noData().End()
}

// ListSubVolumeSnapshots returns a listing of snapshots for a given subvolume.
//
// Similar To:
//  ceph fs subvolume snapshot ls <volume> --group-name=<group> <name>
func (fsa *FSAdmin) ListSubVolumeSnapshots(volume, group, name string) ([]string, error) {
	m := map[string]string{
		"prefix":   "fs subvolume snapshot ls",
		"vol_name": volume,
		"sub_name": name,
		"format":   "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return parseListNames(fsa.marshalMgrCommand(m))
}

// ProtectSubVolumeSnapshot protects the specified snapshot.
//
// Similar To:
//  ceph fs subvolume snapshot protect <volume> --group-name=<group> <subvolume> <name>
func (fsa *FSAdmin) ProtectSubVolumeSnapshot(volume, group, subvolume, name string) error {
	m := map[string]string{
		"prefix":    "fs subvolume snapshot protect",
		"vol_name":  volume,
		"sub_name":  subvolume,
		"snap_name": name,
		"format":    "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return fsa.marshalMgrCommand(m).noData().End()
}

// UnprotectSubVolumeSnapshot removes protection from the specified snapshot.
//
// Similar To:
//  ceph fs subvolume snapshot unprotect <volume> --group-name=<group> <subvolume> <name>
func (fsa *FSAdmin) UnprotectSubVolumeSnapshot(volume, group, subvolume, name string) error {
	m := map[string]string{
		"prefix":    "fs subvolume snapshot unprotect",
		"vol_name":  volume,
		"sub_name":  subvolume,
		"snap_name": name,
		"format":    "json",
	}
	if group != NoGroup {
		m["group_name"] = group
	}
	return fsa.marshalMgrCommand(m).noData().End()
}