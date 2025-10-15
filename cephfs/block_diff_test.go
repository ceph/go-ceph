//go:build ceph_preview

package cephfs

import (
	"crypto/rand"
	"os"
	"path"
	"strings"
	"testing"

	fsadmin "github.com/ceph/go-ceph/cephfs/admin"
	"github.com/ceph/go-ceph/internal/dlsym"
	"github.com/stretchr/testify/assert"
)

func TestFileBlockDiff(t *testing.T) {
	_, cephFileBlockDiffInitErr = dlsym.LookupSymbol("ceph_file_blockdiff_init")
	if cephFileBlockDiffInitErr != nil {
		t.Logf("skipping FileBlockDiff tests: ceph_file_blockdiff_init not found: %v", cephFileBlockDiffInitErr)
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

	subVolPath, err := fsa.SubVolumePath(volume, NoGroup, subname)
	assert.NoError(t, err)
	subVolRootPath := "/volumes/_nogroup/SubVol1"

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	f1name := "file-1.txt"
	f1path := path.Join(subVolPath, f1name)
	relPath := strings.TrimPrefix(f1path, subVolRootPath)

	// 4MB buffer
	randData := make([]byte, 4*1024*1024)
	_, err = rand.Read(randData) // read 4MB of random data
	assert.NoError(t, err)

	// Create a file.
	f1, err := mount.Open(f1path, os.O_RDWR|os.O_CREATE, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)

	// Write 4MB random data to the file.
	_, err = f1.Write(randData)
	assert.NoError(t, err)
	assert.NoError(t, f1.Sync())
	assert.NoError(t, f1.Close())
	defer func() {
		assert.NoError(t, mount.Unlink(f1path))
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

	// Nothing has changed between snap1 and snap2.
	changedBlocks := getChangedBlocks(t, mount, subVolRootPath, relPath, snap1, snap2)
	assert.NotNil(t, changedBlocks)
	assert.Equal(t, 0, len(*changedBlocks))

	_, err = rand.Read(randData) // read 4MB of random data
	assert.NoError(t, err)

	// Write 4MB random data to the file.
	f1, err = mount.Open(f1path, os.O_RDWR, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	_, err = f1.WriteAt(randData, 0)
	assert.NoError(t, err)
	assert.NoError(t, f1.Sync())
	assert.NoError(t, f1.Close())

	snap3 := "Snap3"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap3)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap3)
		assert.NoError(t, err)
	}()

	// 4MB at offset 0 has changed between snap2 and snap3.
	changedBlocks = getChangedBlocks(t, mount, subVolRootPath, relPath, snap2, snap3)
	assert.NotNil(t, changedBlocks)
	assert.Equal(t, len(*changedBlocks), 1)
	assert.Equal(t, (*changedBlocks)[0].Offset, uint64(0))
	assert.Equal(t, (*changedBlocks)[0].Len, uint64(4*1024*1024))

	// read 4MB of random data
	_, err = rand.Read(randData)
	assert.NoError(t, err)

	// Write at multiple intervals to the file.
	f1, err = mount.Open(f1path, os.O_RDWR, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	// write 32 1 byte random data at offset 0 with gap of 20 bytes
	for i := 0; i < 32; i++ {
		_, err = f1.WriteAt(randData[i:i+1], int64(i*21))
		assert.NoError(t, err)
	}
	assert.NoError(t, f1.Sync())
	assert.NoError(t, f1.Close())

	snap4 := "Snap4"
	err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snap4)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snap4)
		assert.NoError(t, err)
	}()

	// test for multiple changed blocks between snap3 and snap4.
	changedBlocks = getChangedBlocks(t, mount, subVolRootPath, relPath, snap3, snap4)
	assert.NotNil(t, changedBlocks)
	assert.Equal(t, len(*changedBlocks), 32)
}

func getChangedBlocks(t *testing.T,
	mount *MountInfo, rootPath, relPath, snap1, snap2 string) *[]ChangedBlock {
	fileBlockDiffInfo, err := FileBlockDiffInit(mount, rootPath, relPath, snap1, snap2)
	assert.NoError(t, err)
	changedBlocksList := make([]ChangedBlock, 0)
	defer func() {
		assert.NoError(t, fileBlockDiffInfo.Close())
	}()
	for {
		fileChangedBlocksInfo, err := fileBlockDiffInfo.Read()
		assert.NoError(t, err)
		if fileChangedBlocksInfo == nil || fileChangedBlocksInfo.NumBlocks == 0 {
			break
		}
		changedBlocksList = append(changedBlocksList, fileChangedBlocksInfo.ChangedBlocks...)
		if !fileBlockDiffInfo.More() {
			break
		}
	}

	return &changedBlocksList
}
