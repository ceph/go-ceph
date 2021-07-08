// +build !nautilus,!octopus

package admin

import (
	"errors"
	"os"
	pth "path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/cephfs"
)

func mirrorConfig() string {
	return os.Getenv("MIRROR_CONF")
}

const (
	noForce      = false
	mirrorClient = "client.mirror_remote"
)

func TestMirroring(t *testing.T) {
	if mirrorConfig() == "" {
		t.Skip("no mirror config available")
	}

	fsa1 := getFSAdmin(t)
	fsname := "cephfs"

	require.NotNil(t, fsa1.conn)
	err := fsa1.EnableMirroringModule(noForce)
	assert.NoError(t, err)
	defer func() {
		err := fsa1.DisableMirroringModule()
		assert.NoError(t, err)
	}()
	require.NoError(t, err)

	smadmin1 := fsa1.SnapshotMirror()
	err = smadmin1.Enable(fsname)
	require.NoError(t, err)
	defer func() {
		err := smadmin1.Disable(fsname)
		require.NoError(t, err)
	}()

	fsa2 := newFSAdmin(t, mirrorConfig())
	err = fsa2.EnableMirroringModule(noForce)
	require.NoError(t, err)
	defer func() {
		err := fsa2.DisableMirroringModule()
		assert.NoError(t, err)
	}()

	smadmin2 := fsa2.SnapshotMirror()
	err = smadmin2.Enable(fsname)
	require.NoError(t, err)
	defer func() {
		err := smadmin2.Disable(fsname)
		require.NoError(t, err)
	}()

	// from https://docs.ceph.com/en/pacific/dev/cephfs-mirroring/
	//  "Peer bootstrap involves creating a bootstrap token on the peer cluster"
	// and "Import the bootstrap token in the primary cluster"
	token, err := smadmin2.CreatePeerBootstrapToken(fsname, mirrorClient, "ceph_b")
	require.NoError(t, err)
	err = smadmin1.ImportPeerBoostrapToken(fsname, token)
	require.NoError(t, err)

	// we need a path to mirror
	path := "/wonderland"

	mount1 := fsConnect(t, "")
	defer func(mount *cephfs.MountInfo) {
		assert.NoError(t, mount.Unmount())
		assert.NoError(t, mount.Release())
	}(mount1)

	mount2 := fsConnect(t, mirrorConfig())
	defer func(mount *cephfs.MountInfo) {
		assert.NoError(t, mount.Unmount())
		assert.NoError(t, mount.Release())
	}(mount2)

	err = mount1.MakeDir(path, 0770)
	require.NoError(t, err)
	defer func() {
		err = mount2.ChangeDir("/")
		assert.NoError(t, err)
		err = mount1.RemoveDir(path)
		assert.NoError(t, err)
	}()
	err = mount2.MakeDir(path, 0770)
	require.NoError(t, err)
	defer func() {
		err = mount2.ChangeDir("/")
		assert.NoError(t, err)
		err = mount2.RemoveDir(path)
		assert.NoError(t, err)
	}()

	err = smadmin1.Add(fsname, path)
	require.NoError(t, err)

	err = mount1.ChangeDir(path)
	require.NoError(t, err)

	// write some dirs & files
	err = mount1.MakeDir("drink_me", 0770)
	require.NoError(t, err)
	err = mount1.MakeDir("eat_me", 0770)
	require.NoError(t, err)
	writeFile(t, mount1, "drink_me/bottle1.txt",
		[]byte("magic potions #1\n"))

	snapname1 := "alice"
	err = mount1.MakeDir(pth.Join(snapDir, snapname1), 0700)
	require.NoError(t, err)
	defer func() {
		err := mount1.RemoveDir(pth.Join(snapDir, snapname1))
		assert.NoError(t, err)
		err = mount2.RemoveDir(pth.Join(snapDir, snapname1))
		assert.NoError(t, err)
	}()

	err = mount2.ChangeDir(path)
	require.NoError(t, err)

	// wait a bit for the snapshot to propagate and the dirs to be created on
	// the remote fs.
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		_, err1 := mount2.Statx("drink_me", cephfs.StatxBasicStats, 0)
		_, err2 := mount2.Statx("eat_me", cephfs.StatxBasicStats, 0)
		if err1 == nil && err2 == nil {
			break
		}
	}

waitforpeers:
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		dstatus, err := smadmin1.DaemonStatus(fsname)
		assert.NoError(t, err)
		for _, dsinfo := range dstatus {
			for _, fsinfo := range dsinfo.FileSystems {
				if len(fsinfo.Peers) > 0 {
					break waitforpeers
				}
			}
		}
	}

	p, err := smadmin1.PeerList(fsname)
	assert.NoError(t, err)
	assert.Len(t, p, 1)
	for _, peer := range p {
		assert.Equal(t, "cephfs", peer.FSName)
	}

	stx, err := mount2.Statx("drink_me", cephfs.StatxBasicStats, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(0040000), stx.Mode&0040000) // is dir?
	}

	stx, err = mount2.Statx("eat_me", cephfs.StatxBasicStats, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(0040000), stx.Mode&0040000) // is dir?
	}

	stx, err = mount2.Statx("drink_me/bottle1.txt", cephfs.StatxBasicStats, 0)
	if assert.NoError(t, err) {
		assert.Equal(t, uint16(0100000), stx.Mode&0100000) // is reg?
		assert.Equal(t, uint64(17), stx.Size)
	}
	data := readFile(t, mount2, "drink_me/bottle1.txt")
	assert.Equal(t, "magic potions #1\n", string(data))

	err = mount1.Unlink("drink_me/bottle1.txt")
	require.NoError(t, err)
	err = mount1.RemoveDir("drink_me")
	require.NoError(t, err)
	err = mount1.RemoveDir("eat_me")
	require.NoError(t, err)

	snapname2 := "rabbit"
	err = mount1.MakeDir(pth.Join(snapDir, snapname2), 0700)
	require.NoError(t, err)
	defer func() {
		err := mount1.RemoveDir(pth.Join(snapDir, snapname2))
		assert.NoError(t, err)
		err = mount2.RemoveDir(pth.Join(snapDir, snapname2))
		assert.NoError(t, err)
	}()

	// wait a bit for the snapshot to propagate and the dirs to be removed on
	// the remote fs.
	for i := 0; i < 60; i++ {
		time.Sleep(500 * time.Millisecond)
		_, err1 := mount2.Statx("drink_me", cephfs.StatxBasicStats, 0)
		_, err2 := mount2.Statx("eat_me", cephfs.StatxBasicStats, 0)
		if err1 != nil && err2 != nil {
			break
		}

	}
	_, err = mount2.Statx("drink_me", cephfs.StatxBasicStats, 0)
	if assert.Error(t, err) {
		var ec errorWithCode
		if assert.True(t, errors.As(err, &ec)) {
			assert.Equal(t, -2, ec.ErrorCode())
		}
	}
	_, err = mount2.Statx("eat_me", cephfs.StatxBasicStats, 0)
	if assert.Error(t, err) {
		var ec errorWithCode
		if assert.True(t, errors.As(err, &ec)) {
			assert.Equal(t, -2, ec.ErrorCode())
		}
	}

	err = smadmin1.Remove(fsname, path)
	assert.NoError(t, err)
}

type errorWithCode interface {
	ErrorCode() int
}
