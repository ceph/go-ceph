package cephfs

import (
	"strings"
	"testing"

	fsadmin "github.com/ceph/go-ceph/cephfs/admin"
	"github.com/stretchr/testify/assert"
)

func TeestFileBlockDiff(t *testing.T) {
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
	t.Logf("path: %v", path)

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	t.Log("getting debug_client=20")
	t.Log(mount.GetConfigOption("debug_client"))

	t.Log("setting debug_client=20", mount.SetConfigOption("debug_client", "20"))
	t.Log("getting debug_client=20")
	t.Log(mount.GetConfigOption("debug_client"))

	t.Log("getting log_file")
	t.Log(mount.GetConfigOption("log_file"))
	t.Log("setting log_file", mount.SetConfigOption("log_file", "/tmp/cephfs.log"))
	t.Log("getting log_file")
	t.Log(mount.GetConfigOption("log_file"))

	err = WriteFile(mount, path, 10)
	assert.NoError(t, err)

	snap1 := "Snap1"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap1)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap1)
		assert.NoError(t, err)
	}()

	err = WriteFile(mount, path, 10)
	assert.NoError(t, err)

	snap2 := "Snap2"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap2)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap2)
		assert.NoError(t, err)
	}()

	t.Log(mount.CurrentDir())
	dirPaths := []string{"/volumes"}
	newDirPaths := []string{}
	for {
		for _, dirPath := range dirPaths {
			t.Logf("dirPath: %v", dirPath)
			Dir, err := mount.OpenDir(dirPath)
			if err != nil {
				t.Logf("open dir %v: %v", dirPath, err)
				continue
			}
			for {
				dirEntry, err := Dir.ReadDir()
				if err != nil {
					t.Log(err)
					continue
				}
				if dirEntry == nil {
					break
				}
				if dirEntry.Name() == "." || dirEntry.Name() == ".." {
					continue
				}
				t.Logf("dirEntry: %v: %v: %v", dirEntry.Name(), dirEntry.Inode(), dirEntry.DType())
				if dirEntry.DType() == DTypeDir {
					newDirPaths = append(newDirPaths, dirPath+"/"+dirEntry.Name())
				}
			}
		}
		if len(newDirPaths) == 0 {
			break
		}
		dirPaths = newDirPaths
		newDirPaths = []string{}
	}

	snap1ID, err := GetSnapshotID(mount, "/volumes/_nogroup/SubVol1/.snap/"+snap1)
	assert.NoError(t, err)
	snap2ID, err := GetSnapshotID(mount, "/volumes/_nogroup/SubVol1/.snap/"+snap2)
	assert.NoError(t, err)
	t.Logf("snap1ID: %v", snap1ID)
	t.Logf("snap2ID: %v", snap2ID)
	err = mount.ChangeDir("/")
	assert.NoError(t, err)
	t.Log(mount.CurrentDir())
	t.Log(path)

	splits := strings.Split(path, "/")
	relPath := splits[len(splits)-1] + "/file-0.txt"
	t.Log(splits)
	t.Log(relPath)
	cephFileBlockDiffInfo, err := CephFileBlockDiffInit(mount, "/volumes/_nogroup/SubVol1/", relPath, snap1, snap2)
	assert.NoError(t, err)
	for {
		fileChangedBlocksInfo, err := CephFileBlockDiff(cephFileBlockDiffInfo)
		t.Log(err)
		if err != nil {
			t.Logf("CephFileBlockDiff: %v", err)
			break
		}
		if fileChangedBlocksInfo == nil {
			break
		}
		t.Logf("fileChangedBlocksInfo: %v", fileChangedBlocksInfo)
		for i, Cblock := range fileChangedBlocksInfo.CBlocks {
			t.Logf("Cblock[%d]: offset: %d, len: %d", i, Cblock.Offset, Cblock.Len)
		}
	}

	err = CephFileBlockDiffFinish(cephFileBlockDiffInfo)
	t.Log("CephFileBlockDiffFinish: ", err)
}
