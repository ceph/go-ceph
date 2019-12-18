package cephfs

import (
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	CephMountTest = "/tmp/ceph/mds/mnt/"
)

func TestCreateMount(t *testing.T) {
	mount := fsConnect(t)
	mount, err := CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
}

func fsConnect(t *testing.T) *MountInfo {
	mount, err := CreateMount()
	require.NoError(t, err)
	require.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	require.NoError(t, err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(mount *MountInfo) {
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

func TestMountRoot(t *testing.T) {
	fsConnect(t)
}

func TestSyncFs(t *testing.T) {
	mount := fsConnect(t)

	err := mount.SyncFs()
	assert.NoError(t, err)
}

func TestChangeDir(t *testing.T) {
	mount := fsConnect(t)

	dir1 := mount.CurrentDir()
	assert.NotNil(t, dir1)

	err := mount.MakeDir("/asdf", 0755)
	assert.NoError(t, err)

	err = mount.ChangeDir("/asdf")
	assert.NoError(t, err)

	dir2 := mount.CurrentDir()
	assert.NotNil(t, dir2)

	assert.NotEqual(t, dir1, dir2)
	assert.Equal(t, dir1, "/")
	assert.Equal(t, dir2, "/asdf")
}

func TestRemoveDir(t *testing.T) {
	dirname := "one"
	mount := fsConnect(t)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	_, err = os.Stat(CephMountTest + dirname)
	assert.NoError(t, err)

	err = mount.RemoveDir(dirname)
	assert.NoError(t, err)

	_, err = os.Stat(CephMountTest + dirname)
	assert.EqualError(t, err,
		fmt.Sprintf("stat %s: no such file or directory", CephMountTest+dirname))
}

func TestUnmountMount(t *testing.T) {
	t.Run("neverMounted", func(t *testing.T) {
		mount, err := CreateMount()
		require.NoError(t, err)
		require.NotNil(t, mount)
		assert.False(t, mount.IsMounted())
	})
	t.Run("mountUnmount", func(t *testing.T) {
		mount := fsConnect(t)
		assert.True(t, mount.IsMounted())

		err := mount.Unmount()
		assert.NoError(t, err)
		assert.False(t, mount.IsMounted())
	})
}

func TestReleaseMount(t *testing.T) {
	mount, err := CreateMount()
	assert.NoError(t, err)
	require.NotNil(t, mount)

	err = mount.Release()
	assert.NoError(t, err)
}

func TestChmodDir(t *testing.T) {
	dirname := "two"
	var stats_before uint32 = 0755
	var stats_after uint32 = 0700
	mount := fsConnect(t)

	err := mount.MakeDir(dirname, stats_before)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(CephMountTest + dirname)
	require.NoError(t, err)

	assert.Equal(t, uint32(stats.Mode().Perm()), stats_before)

	err = mount.Chmod(dirname, stats_after)
	assert.NoError(t, err)

	stats, err = os.Stat(CephMountTest + dirname)
	assert.Equal(t, uint32(stats.Mode().Perm()), stats_after)
}

// Not cross-platform, go's os does not specifiy Sys return type
func TestChown(t *testing.T) {
	dirname := "three"
	// dockerfile creates bob user account
	var bob uint32 = 1010
	var root uint32

	mount := fsConnect(t)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(CephMountTest + dirname)
	require.NoError(t, err)

	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), root)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), root)

	err = mount.Chown(dirname, bob, bob)
	assert.NoError(t, err)

	stats, err = os.Stat(CephMountTest + dirname)
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), bob)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), bob)

}

func TestCephFSError(t *testing.T) {
	err := getError(0)
	assert.NoError(t, err)

	err = getError(-5) // IO error
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "cephfs: ret=5, Input/output error")

	err = getError(345) // no such errno
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "cephfs: ret=345")
}

func radosConnect(t *testing.T) *rados.Conn {
	conn, err := rados.NewConn()
	require.NoError(t, err)
	err = conn.ReadDefaultConfigFile()
	require.NoError(t, err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(conn *rados.Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = fmt.Errorf("timed out waiting for connect")
	}
	require.NoError(t, err)
	return conn
}

func TestCreateFromRados(t *testing.T) {
	conn := radosConnect(t)
	mount, err := CreateFromRados(conn)
	assert.NoError(t, err)
	assert.NotNil(t, mount)
}
