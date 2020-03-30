package cephfs

import (
	"encoding/json"
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

func TestCreateMountWithId(t *testing.T) {
	mount, err := CreateMountWithId("bobolink")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)

	// verify the custom entity_id is visible in the 'session ls' output
	// of mds.
	cmd := []byte(`{"prefix": "session ls"}`)
	buf, info, err := mount.MdsCommand(
		"Z", // TODO: fix hard-coded name mds (from ci container script)
		[][]byte{cmd})
	assert.NoError(t, err)
	assert.NotEqual(t, "", string(buf))
	assert.Equal(t, "", string(info))
	assert.Contains(t, string(buf), `"bobolink"`)
}

func TestMdsCommand(t *testing.T) {
	mount := fsConnect(t)

	cmd := []byte(`{"prefix": "client ls"}`)
	buf, info, err := mount.MdsCommand(
		"Z", // TODO: fix hard-coded name mds (from ci container script)
		[][]byte{cmd})
	assert.NoError(t, err)
	assert.NotEqual(t, "", string(buf))
	assert.Equal(t, "", string(info))
	assert.Contains(t, string(buf), "ceph_version")
	// response should also be valid json
	var j []interface{}
	err = json.Unmarshal(buf, &j)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(j), 1)
}

func TestMdsCommandError(t *testing.T) {
	mount := fsConnect(t)

	cmd := []byte("iAMinValId~~~")
	buf, info, err := mount.MdsCommand(
		"Z", // TODO: fix hard-coded name mds (from ci container script)
		[][]byte{cmd})
	assert.Error(t, err)
	assert.Equal(t, "", string(buf))
	assert.NotEqual(t, "", string(info))
	assert.Contains(t, string(info), "unparseable JSON")
}

func TestMountWithRoot(t *testing.T) {
	bMount := fsConnect(t)
	defer func() {
		assert.NoError(t, bMount.Unmount())
		assert.NoError(t, bMount.Release())
	}()

	dir1 := "/test-mount-with-root"
	err := bMount.MakeDir(dir1, 0755)
	assert.NoError(t, err)
	defer bMount.RemoveDir(dir1)

	sub1 := "/i.was.here"
	dir2 := dir1 + sub1
	err = bMount.MakeDir(dir2, 0755)
	assert.NoError(t, err)
	defer bMount.RemoveDir(dir2)

	t.Run("withRoot", func(t *testing.T) {
		mount, err := CreateMount()
		require.NoError(t, err)
		require.NotNil(t, mount)
		defer func() {
			assert.NoError(t, mount.Unmount())
			assert.NoError(t, mount.Release())
		}()

		err = mount.ReadDefaultConfigFile()
		require.NoError(t, err)

		err = mount.MountWithRoot(dir1)
		assert.NoError(t, err)

		err = mount.ChangeDir(sub1)
		assert.NoError(t, err)
	})
	t.Run("badRoot", func(t *testing.T) {
		mount, err := CreateMount()
		require.NoError(t, err)
		require.NotNil(t, mount)
		defer func() {
			assert.NoError(t, mount.Release())
		}()

		err = mount.ReadDefaultConfigFile()
		require.NoError(t, err)

		err = mount.MountWithRoot("/i-yam-what-i-yam")
		assert.Error(t, err)
	})
}

func TestGetSetConfigOption(t *testing.T) {
	// we don't need an active connection for this, just the handle
	mount, err := CreateMount()
	require.NoError(t, err)
	require.NotNil(t, mount)

	err = mount.SetConfigOption("__dne__", "value")
	assert.Error(t, err)
	_, err = mount.GetConfigOption("__dne__")
	assert.Error(t, err)

	origVal, err := mount.GetConfigOption("log_file")
	assert.NoError(t, err)

	err = mount.SetConfigOption("log_file", "/dev/null")
	assert.NoError(t, err)
	currVal, err := mount.GetConfigOption("log_file")
	assert.NoError(t, err)
	assert.Equal(t, "/dev/null", currVal)

	err = mount.SetConfigOption("log_file", origVal)
	assert.NoError(t, err)
	currVal, err = mount.GetConfigOption("log_file")
	assert.NoError(t, err)
	assert.Equal(t, origVal, currVal)
}
