package cephfs_test

import (
	"fmt"
	"github.com/ceph/go-ceph/cephfs"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestCreateMount(t *testing.T) {
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())
}

func TestMountRoot(t *testing.T) {
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)
}

func TestSyncFs(t *testing.T) {
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)
}

func TestChangeDir(t *testing.T) {
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)

	dir1 := mount.CurrentDir()
	assert.NotNil(t, dir1)

	err = mount.MakeDir("/asdf", 0755)
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
	dirname := "/one"
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)

	dir1 := mount.CurrentDir()
	assert.NotNil(t, dir1)

	fmt.Printf("path: %v\n", dir1)

	err = mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	files, _ := ioutil.ReadDir("./")
	for _, f := range files {
		fmt.Println(f.Name())
	}

	_, err = os.Stat(dirname)
	assert.NoError(t, err)

	err = mount.RemoveDir(dirname)
	assert.NoError(t, err)
	_, err = os.Stat(dirname)
	assert.EqualError(t, err, fmt.Sprintf("stat %s: no such file or directory", dirname))
}

func TestUnmountMount(t *testing.T) {
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())

	err = mount.Unmount()
	assert.NoError(t, err)
	assert.False(t, mount.IsMounted())
}

func TestReleaseMount(t *testing.T) {
	mount, err := cephfs.CreateMount()
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	assert.True(t, mount.IsMounted())

	err = mount.Release()
	assert.NoError(t, err)
	assert.Nil(t, mount.GetMount())
}
