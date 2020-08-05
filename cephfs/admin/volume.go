// +build !luminous,!mimic

package admin

var listVolumesCmd = []byte(`{"prefix":"fs volume ls"}`)

// ListVolumes return a list of volumes in this Ceph cluster.
func (fsa *FSAdmin) ListVolumes() ([]string, error) {
	r, s, err := fsa.rawMgrCommand(listVolumesCmd)
	return parseListNames(r, s, err)
}
