package cephfs

import (
	"os"
	"strings"
	"testing"

	fsadmin "github.com/ceph/go-ceph/cephfs/admin"
	"github.com/ceph/go-ceph/internal/admintest"
	"github.com/ceph/go-ceph/internal/dlsym"
	"github.com/stretchr/testify/assert"
)

var (
	radosConnector = admintest.NewConnector()
)

// NoGroup should be used when an optional subvolume group name is not
// specified.
const NoGroup = ""

func TestSnapDiff(t *testing.T) {
	_, cephOpenSnapDiffErr := dlsym.LookupSymbol("ceph_open_snapdiff")
	if cephOpenSnapDiffErr != nil {
		t.Logf("skipping SnapDiff tests: ceph_open_snapdiff not found: %v", cephOpenSnapDiffErr)

		return
	}

	fsa := fsadmin.NewFromConn(radosConnector.Get(t))
	volume := "cephfs"

	subname := "SubVol1"
	err := fsa.CreateSubVolume(volume, NoGroup, subname, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, NoGroup, subname)
		assert.NoError(t, err)
	}()

	path, err := fsa.SubVolumePath(volume, NoGroup, subname)
	assert.NoError(t, err)
	subVolRootPath := "/volumes/_nogroup/SubVol1"
	relPath := strings.TrimPrefix(path, subVolRootPath)

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	f1name := path + "/file-1.txt"
	f1, err := mount.Open(f1name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	err = f1.Close()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, mount.Unlink(f1name))
	}()

	snap1 := "Snap1"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap1)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap1)
		assert.NoError(t, err)
	}()

	snap2 := "Snap2"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap2)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap2)
		assert.NoError(t, err)
	}()

	changedFiles := getChangedFiles(t, SnapDiffConfig{
		CMount:   mount,
		RootPath: subVolRootPath,
		RelPath:  relPath,
		Snap1:    snap1,
		Snap2:    snap2,
	})
	// No changes between the two snapshots.
	assert.Len(t, changedFiles, 0)

	f2name := path + "/file-2.txt"
	f2, err := mount.Open(f2name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f2)
	err = f2.Close()
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, mount.Unlink(f2name))
	}()

	snap3 := "Snap3"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap3)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap3)
		assert.NoError(t, err)
	}()

	changedFiles = getChangedFiles(t, SnapDiffConfig{
		CMount:   mount,
		RootPath: subVolRootPath,
		RelPath:  relPath,
		Snap1:    snap2,
		Snap2:    snap3,
	})
	// one changed file between the two snapshots.
	assert.Len(t, changedFiles, 1)
	assert.NotEqual(t, changedFiles[0], f2name)
}

func getChangedFiles(t *testing.T, snapDiffConfig SnapDiffConfig) []string {
	diff, err := CephOpenSnapDiff(snapDiffConfig)
	assert.NoError(t, err)
	assert.NotNil(t, diff)
	assert.NotNil(t, diff.CMount)
	assert.NotNil(t, diff.Dir1)
	assert.NotNil(t, diff.DirAux)
	defer func() {
		assert.NoError(t, CephCloseSnapDiff(diff))
	}()

	changedFiles := []string{}
	for {
		diffEntry, err := CephReaddirSnapDiff(diff)
		if err != nil {
			t.Errorf("readdir snap diff error: %v", err)
			assert.NoError(t, err)
		}
		if diffEntry == nil {
			break
		}
		if diffEntry.DirEntry.Name() == "." || diffEntry.DirEntry.Name() == ".." {
			continue
		}
		changedFiles = append(changedFiles, diffEntry.DirEntry.Name())
	}

	return changedFiles
}
