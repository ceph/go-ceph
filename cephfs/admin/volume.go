// +build !luminous,!mimic

package admin

var (
	listVolumesCmd = []byte(`{"prefix":"fs volume ls"}`)
	dumpVolumesCmd = []byte(`{"prefix":"fs dump","format":"json"}`)
)

// ListVolumes return a list of volumes in this Ceph cluster.
//
// Similar To:
//  ceph fs volume ls
func (fsa *FSAdmin) ListVolumes() ([]string, error) {
	r, s, err := fsa.rawMgrCommand(listVolumesCmd)
	return parseListNames(r, s, err)
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
)

func parseDumpToIdents(r []byte, s string, err error) ([]VolumeIdent, error) {
	if len(s) >= dumpOkLen && s[:dumpOkLen] == dumpOkPrefix {
		// Unhelpfully, ceph drops a status string on success responses for this
		// call. this hacks around that by ignoring its typical prefix
		s = ""
	}
	var dump fsDump
	if err := unmarshalResponseJSON(r, s, err, &dump); err != nil {
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
	r, s, err := fsa.rawMonCommand(dumpVolumesCmd)
	return parseDumpToIdents(r, s, err)
}
