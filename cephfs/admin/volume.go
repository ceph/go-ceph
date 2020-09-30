// +build !luminous,!mimic

package admin

import (
	"bytes"
)

var (
	listVolumesCmd = []byte(`{"prefix":"fs volume ls"}`)
	dumpVolumesCmd = []byte(`{"prefix":"fs dump","format":"json"}`)
)

// ListVolumes return a list of volumes in this Ceph cluster.
//
// Similar To:
//  ceph fs volume ls
func (fsa *FSAdmin) ListVolumes() ([]string, error) {
	res := fsa.rawMgrCommand(listVolumesCmd)
	return parseListNames(res)
}

// VolumeIdent contains a pair of file system identifying values: the volume
// name and the volume ID.
type VolumeIdent struct {
	Name string
	ID   int64
}

type cephFileSystem struct {
	ID     int64 `json:"id"`
	MDSMap struct {
		FSName string `json:"fs_name"`
	} `json:"mdsmap"`
}

type fsDump struct {
	FileSystems []cephFileSystem `json:"filesystems"`
}

const (
	dumpOkPrefix = "dumped fsmap epoch"
	dumpOkLen    = len(dumpOkPrefix)

	invalidTextualResponse = "this ceph version returns a non-parsable volume status response"
)

func parseDumpToIdents(res response) ([]VolumeIdent, error) {
	if !res.Ok() {
		return nil, res.End()
	}
	if len(res.status) >= dumpOkLen && res.status[:dumpOkLen] == dumpOkPrefix {
		// Unhelpfully, ceph drops a status string on success responses for this
		// call. this hacks around that by ignoring its typical prefix
		res.status = ""
	}
	var dump fsDump
	if err := res.noStatus().unmarshal(&dump).End(); err != nil {
		return nil, err
	}
	// copy the dump json into the simpler enumeration list
	idents := make([]VolumeIdent, len(dump.FileSystems))
	for i := range dump.FileSystems {
		idents[i].ID = dump.FileSystems[i].ID
		idents[i].Name = dump.FileSystems[i].MDSMap.FSName
	}
	return idents, nil
}

// EnumerateVolumes returns a list of volume-name volume-id pairs.
func (fsa *FSAdmin) EnumerateVolumes() ([]VolumeIdent, error) {
	// We base our enumeration on the ceph fs dump json.  This may not be the
	// only way to do it, but it's the only one I know of currently. Because of
	// this and to keep our initial implementation simple, we expose our own
	// simplified type only, rather do a partial implementation of dump.
	return parseDumpToIdents(fsa.rawMonCommand(dumpVolumesCmd))
}

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
	res = res.noStatus()
	if !res.Ok() {
		return nil, res.End()
	}
	res = res.unmarshal(&vs)
	if !res.Ok() {
		if bytes.HasPrefix(res.body, []byte("ceph")) {
			res.status = invalidTextualResponse
			return nil, NotImplementedError{response: res}
		}
		return nil, res.End()
	}
	return &vs, nil
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
