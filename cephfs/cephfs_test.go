package cephfs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testMdsName = "Z"
)

func init() {
	mdsName := os.Getenv("GO_CEPH_TEST_MDS_NAME")
	if mdsName != "" {
		testMdsName = mdsName
	}
}

func TestCreateMount(t *testing.T) {
	mount, err := CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.NoError(t, mount.Release())
}

func fsConnect(t require.TestingT) *MountInfo {
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

func fsDisconnect(t assert.TestingT, mount *MountInfo) {
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

func TestParseConfigArgv(t *testing.T) {
	mount, err := CreateMount()
	require.NoError(t, err)
	require.NotNil(t, mount)
	defer func() { assert.NoError(t, mount.Release()) }()

	origVal, err := mount.GetConfigOption("log_file")
	assert.NoError(t, err)

	err = mount.ParseConfigArgv(
		[]string{"cephfs.test", "--log_file", "/dev/null"})
	assert.NoError(t, err)

	currVal, err := mount.GetConfigOption("log_file")
	assert.NoError(t, err)
	assert.Equal(t, "/dev/null", currVal)
	assert.NotEqual(t, "/dev/null", origVal)

	// ensure that an empty slice triggers an error (not a crash)
	err = mount.ParseConfigArgv([]string{})
	assert.Error(t, err)

	// ensure we get an error for an invalid mount value
	badMount := &MountInfo{}
	err = badMount.ParseConfigArgv(
		[]string{"cephfs.test", "--log_file", "/dev/null"})
	assert.Error(t, err)
}

func TestParseDefaultConfigEnv(t *testing.T) {
	mount, err := CreateMount()
	require.NoError(t, err)
	require.NotNil(t, mount)
	defer func() { assert.NoError(t, mount.Release()) }()

	origVal, err := mount.GetConfigOption("log_file")
	assert.NoError(t, err)

	err = os.Setenv("CEPH_ARGS", "--log_file /dev/null")
	assert.NoError(t, err)
	err = mount.ParseDefaultConfigEnv()
	assert.NoError(t, err)

	currVal, err := mount.GetConfigOption("log_file")
	assert.NoError(t, err)
	assert.Equal(t, "/dev/null", currVal)
	assert.NotEqual(t, "/dev/null", origVal)
}

func TestValidate(t *testing.T) {
	mount, err := CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	defer assert.NoError(t, mount.Release())

	t.Run("mountCurrentDir", func(t *testing.T) {
		path := mount.CurrentDir()
		assert.Equal(t, path, "")
	})

	t.Run("mountChangeDir", func(t *testing.T) {
		err := mount.ChangeDir("someDir")
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})

	t.Run("mountMakeDir", func(t *testing.T) {
		err := mount.MakeDir("someName", 0444)
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})

	t.Run("mountRemoveDir", func(t *testing.T) {
		err := mount.RemoveDir("someDir")
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})

	t.Run("mountLink", func(t *testing.T) {
		err := mount.Link("/", "/")
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})

	t.Run("mountUnlink", func(t *testing.T) {
		err := mount.Unlink("someFile")
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})

	t.Run("mountSymlink", func(t *testing.T) {
		err := mount.Symlink("/", "/")
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})

	t.Run("mountReadlink", func(t *testing.T) {
		_, err := mount.Readlink("somePath")
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotConnected)
	})
}
