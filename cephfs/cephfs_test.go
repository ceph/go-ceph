package cephfs

import (
	"fmt"
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	CephMountDir     = "/tmp/ceph/mds/mnt/"
	requireCephMount = false
	testMdsName      = "Z"
)

func init() {
	mdir := os.Getenv("GO_CEPH_TEST_MOUNT_DIR")
	if mdir != "" {
		CephMountDir = mdir
	}
	reqMount := os.Getenv("GO_CEPH_TEST_REQUIRE_MOUNT")
	if reqMount == "yes" || reqMount == "true" {
		requireCephMount = true
	}
	mdsName := os.Getenv("GO_CEPH_TEST_MDS_NAME")
	if mdsName != "" {
		testMdsName = mdsName
	}
}

func useMount(t *testing.T) {
	fail := func(m string) {
		if requireCephMount {
			t.Fatalf("cephfs mount required: %s %s", CephMountDir, m)
		} else {
			t.Skipf("cephfs mount needed: %s %s", CephMountDir, m)
		}
	}

	s, err := os.Stat(CephMountDir)
	if err != nil || !s.IsDir() {
		fail("missing or not a directory")
	}

	if us, ok := s.Sys().(*syscall.Stat_t); ok {
		ps, err := os.Stat(path.Dir(path.Clean(CephMountDir)))
		if err != nil {
			fail("missing parent directory (race condition?)")
		}
		if ps.Sys().(*syscall.Stat_t).Dev == us.Dev {
			fail("not a mount point")
		}
	} else {
		fail("not a unix-like file system? how did you even compile this?" +
			"no, seriously please contact us or file an issue and let us know!")
	}
}

func TestCreateMount(t *testing.T) {
	mount, err := CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.NoError(t, mount.Release())
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

func fsDisconnect(t *testing.T, mount *MountInfo) {
	assert.NoError(t, mount.Unmount())
	assert.NoError(t, mount.Release())
}

func TestMountRoot(t *testing.T) {
	mount := fsConnect(t)
	fsDisconnect(t, mount)
}

func TestSyncFs(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.SyncFs()
	assert.NoError(t, err)
}

func TestUnmountMount(t *testing.T) {
	t.Run("neverMounted", func(t *testing.T) {
		mount, err := CreateMount()
		require.NoError(t, err)
		require.NotNil(t, mount)
		assert.False(t, mount.IsMounted())
		assert.NoError(t, mount.Release())
	})
	t.Run("mountUnmount", func(t *testing.T) {
		mount := fsConnect(t)
		defer func() { assert.NoError(t, mount.Release()) }()
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

	assert.NoError(t, mount.Release())
	// call release again to ensure idempotency of the func
	assert.NoError(t, mount.Release())
}

func TestChmodDir(t *testing.T) {
	useMount(t)

	dirname := "two"
	var stats_before uint32 = 0755
	var stats_after uint32 = 0700
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, stats_before)
	assert.NoError(t, err)
	defer mount.RemoveDir(dirname)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(path.Join(CephMountDir, dirname))
	require.NoError(t, err)

	assert.Equal(t, uint32(stats.Mode().Perm()), stats_before)

	err = mount.Chmod(dirname, stats_after)
	assert.NoError(t, err)

	stats, err = os.Stat(path.Join(CephMountDir, dirname))
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Mode().Perm()), stats_after)
}

// Not cross-platform, go's os does not specifiy Sys return type
func TestChown(t *testing.T) {
	useMount(t)

	dirname := "three"
	// dockerfile creates bob user account
	var bob uint32 = 1010
	var root uint32

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)
	defer mount.RemoveDir(dirname)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(path.Join(CephMountDir, dirname))
	require.NoError(t, err)

	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), root)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), root)

	err = mount.Chown(dirname, bob, bob)
	assert.NoError(t, err)

	stats, err = os.Stat(path.Join(CephMountDir, dirname))
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
	defer func() { assert.NoError(t, mount.Release()) }()

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.Unmount()) }()

	// verify the custom entity_id is visible in the 'session ls' output
	// of mds.
	cmd := []byte(`{"prefix": "session ls"}`)
	buf, info, err := mount.MdsCommand(
		testMdsName,
		[][]byte{cmd})
	assert.NoError(t, err)
	assert.NotEqual(t, "", string(buf))
	assert.Equal(t, "", string(info))
	assert.Contains(t, string(buf), `"bobolink"`)
}

func TestMountWithRoot(t *testing.T) {
	bMount := fsConnect(t)
	defer fsDisconnect(t, bMount)

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
			assert.NoError(t, mount.Release())
		}()

		err = mount.ReadDefaultConfigFile()
		require.NoError(t, err)

		err = mount.MountWithRoot(dir1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, mount.Unmount())
		}()

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
	defer func() { assert.NoError(t, mount.Release()) }()

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
