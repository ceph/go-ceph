//go:build !luminous && !mimic
// +build !luminous,!mimic

package admin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	pathpkg "path"
	"testing"
	"time"

	"github.com/ceph/go-ceph/cephfs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var snapDir = ".snapshots"

func fsConnect(t *testing.T, configFile string) *cephfs.MountInfo {
	mount, err := cephfs.CreateMount()
	require.NoError(t, err)
	require.NotNil(t, mount)

	if configFile == "" {
		err = mount.ReadDefaultConfigFile()
		require.NoError(t, err)
	} else {
		err = mount.ReadConfigFile(configFile)
		require.NoError(t, err)
	}
	err = mount.SetConfigOption("client_snapdir", snapDir)
	require.NoError(t, err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(mount *cephfs.MountInfo) {
		ch <- mount.Mount()
	}(mount)
	select {
	case err = <-ch:
	case <-timeout:
		err = fmt.Errorf("timed out waiting for connect")
	}
	require.NoError(t, err)
	return mount
}

func writeFile(t *testing.T, mount *cephfs.MountInfo, path string, content []byte) {
	f1, err := mount.Open(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	require.NoError(t, err)
	defer f1.Close()
	f1.WriteAt(content, 0)
}

func readFile(t *testing.T, mount *cephfs.MountInfo, path string) []byte {
	f1, err := mount.Open(path, os.O_RDONLY, 0)
	require.NoError(t, err)
	defer f1.Close()
	b, err := ioutil.ReadAll(f1)
	require.NoError(t, err)
	return b
}

func getSnapPath(t *testing.T, mount *cephfs.MountInfo, subvol, snapname string) string {
	// I wish there was a nicer way to do this
	snapPath := pathpkg.Join(subvol, snapDir, snapname)
	_, err := mount.Statx(snapPath, cephfs.StatxBasicStats, 0)
	if err == nil {
		return snapPath
	}
	snapPath = pathpkg.Join(
		pathpkg.Dir(subvol),
		snapDir,
		snapname,
		pathpkg.Base(subvol))
	_, err = mount.Statx(snapPath, cephfs.StatxBasicStats, 0)
	if err == nil {
		return snapPath
	}
	t.Fatalf("did not find a snap path for %s", snapname)
	return ""
}

// TestWorkflow aims to do more than just exercise the API calls, but rather to
// also check that they do what they say on the tin.  This means importing the
// cephfs lib in addition to the admin lib and reading and writing to the
// subvolume, snapshot, and clone as appropriate.
func TestWorkflow(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "workflow1"

	// verify the volume we want to use
	l, err := fsa.ListVolumes()
	require.NoError(t, err)
	require.Contains(t, l, volume)

	err = fsa.CreateSubVolumeGroup(volume, group, nil)
	require.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	subname := "files1"
	svopts := &SubVolumeOptions{
		Mode: 0777,
		Size: 2 * gibiByte,
	}
	err = fsa.CreateSubVolume(volume, group, subname, svopts)
	require.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, subname)
		assert.NoError(t, err)
	}()

	// getpath
	subPath, err := fsa.SubVolumePath(volume, group, subname)
	require.NoError(t, err)
	require.NotEqual(t, "", subPath)

	// connect to volume, cd to path (?)
	mount := fsConnect(t, "")
	defer func(mount *cephfs.MountInfo) {
		assert.NoError(t, mount.Unmount())
		assert.NoError(t, mount.Release())
	}(mount)

	err = mount.ChangeDir(subPath)
	require.NoError(t, err)

	// write some dirs & files
	err = mount.MakeDir("content1", 0770)
	require.NoError(t, err)

	writeFile(t, mount, "content1/robots.txt",
		[]byte("robbie\nr2\nbender\nclaptrap\n"))
	writeFile(t, mount, "content1/songs.txt",
		[]byte("none of them knew they were robots\n"))
	assert.NoError(t, mount.MakeDir("content1/emptyDir1", 0770))

	err = mount.MakeDir("content2", 0770)
	require.NoError(t, err)

	writeFile(t, mount, "content2/androids.txt",
		[]byte("data\nmarvin\n"))
	assert.NoError(t, mount.MakeDir("content2/docs", 0770))
	writeFile(t, mount, "content2/docs/lore.txt",
		[]byte("Compendium\nLegend\nLore\nDeadweight\nSpirit at Aphelion\n"))

	assert.NoError(t, mount.SyncFs())

	// take a snapshot

	snapname1 := "hotSpans1"
	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname1)
	require.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
	}()

	sinfo, err := fsa.SubVolumeSnapshotInfo(volume, group, subname, snapname1)
	require.NoError(t, err)
	require.NotNil(t, sinfo)

	time.Sleep(500 * time.Millisecond) // is there a race?

	// examine the snapshot
	snapPath := getSnapPath(t, mount, subPath, snapname1)
	require.NotEqual(t, "", snapPath)

	tempPath := pathpkg.Join(snapPath, "content1/robots.txt")
	txt := readFile(t, mount, tempPath)
	assert.Contains(t, string(txt), "robbie")

	// original subvol can be manipulated
	err = mount.Rename("content2/docs/lore.txt", "content1/lore.txt")
	assert.NoError(t, err)
	writeFile(t, mount, "content1/songs.txt",
		[]byte("none of them knew they were robots\nars moriendi\n"))

	// snapshot may not be modified
	err = mount.Rename(
		pathpkg.Join(snapPath, "content2/docs/lore.txt"),
		pathpkg.Join(snapPath, "content1/lore.txt"))
	assert.Error(t, err)
	txt = readFile(t, mount, pathpkg.Join(snapPath, "content2/docs/lore.txt"))
	assert.Contains(t, string(txt), "Spirit")

	// make a clone

	clonename := "files2"
	err = fsa.CloneSubVolumeSnapshot(
		volume, group, subname, snapname1, clonename,
		&CloneOptions{TargetGroup: group})
	var x NotProtectedError
	if errors.As(err, &x) {
		err = fsa.ProtectSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
		defer func() {
			err := fsa.UnprotectSubVolumeSnapshot(volume, group, subname, snapname1)
			assert.NoError(t, err)
		}()

		err = fsa.CloneSubVolumeSnapshot(
			volume, group, subname, snapname1, clonename,
			&CloneOptions{TargetGroup: group})
	}
	require.NoError(t, err)
	defer func() {
		err := fsa.ForceRemoveSubVolume(volume, group, clonename)
		assert.NoError(t, err)
	}()

	// wait for cloning to complete
	for done := false; !done; {
		status, err := fsa.CloneStatus(volume, group, clonename)
		assert.NoError(t, err)
		assert.NotNil(t, status)
		switch status.State {
		case ClonePending, CloneInProgress:
			time.Sleep(5 * time.Millisecond)
		case CloneComplete:
			done = true
		case CloneFailed:
			t.Fatal("clone failed")
		default:
			t.Fatalf("invalid status.State: %q", status.State)
		}
	}

	// examine the clone
	clonePath, err := fsa.SubVolumePath(volume, group, clonename)
	require.NoError(t, err)
	require.NotEqual(t, "", clonePath)

	txt = readFile(t, mount, pathpkg.Join(clonePath, "content1/robots.txt"))
	assert.Contains(t, string(txt), "robbie")

	// clones are r/w
	err = mount.Rename(
		pathpkg.Join(clonePath, "content2/docs/lore.txt"),
		pathpkg.Join(clonePath, "content1/lore.txt"))
	assert.NoError(t, err)
	txt = readFile(t, mount, pathpkg.Join(clonePath, "content1/lore.txt"))
	assert.Contains(t, string(txt), "Spirit")

	// it reflects what was in the snapshot
	txt = readFile(t, mount, pathpkg.Join(clonePath, "content1/songs.txt"))
	assert.Contains(t, string(txt), "robots")
	assert.NotContains(t, string(txt), "moriendi")

	// ... with it's own independent data
	writeFile(t, mount, pathpkg.Join(clonePath, "content1/songs.txt"),
		[]byte("none of them knew they were robots\nsweet charity\n"))

	// (orig)
	txt = readFile(t, mount, "content1/songs.txt")
	assert.Contains(t, string(txt), "robots")
	assert.Contains(t, string(txt), "moriendi")
	assert.NotContains(t, string(txt), "charity")

	// (clone)
	txt = readFile(t, mount, pathpkg.Join(clonePath, "content1/songs.txt"))
	assert.Contains(t, string(txt), "robots")
	assert.NotContains(t, string(txt), "moriendi")
	assert.Contains(t, string(txt), "charity")
}
