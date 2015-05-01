package cephfs_test

import "testing"
import "github.com/noahdesu/go-ceph/cephfs"
import "github.com/stretchr/testify/assert"

func TestCreateMount(t *testing.T) {
    mount, err := cephfs.CreateMount()
    assert.NoError(t, err)
    assert.NotNil(t, mount)
}

func TestMountRoot(t *testing.T) {
    mount, err := cephfs.CreateMount()
    assert.NoError(t, err)
    assert.NotNil(t, mount)

    err = mount.ReadDefaultConfigFile()
    assert.NoError(t, err)

    err = mount.Mount()
    assert.NoError(t, err)
}

func TestSyncFs(t *testing.T) {
    mount, err := cephfs.CreateMount()
    assert.NoError(t, err)
    assert.NotNil(t, mount)

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

    err = mount.ReadDefaultConfigFile()
    assert.NoError(t, err)

    err = mount.Mount()
    assert.NoError(t, err)

    err = mount.ChangeDir("/")
    assert.NoError(t, err)
}

func TestCurrentDir(t *testing.T) {
    mount, err := cephfs.CreateMount()
    assert.NoError(t, err)
    assert.NotNil(t, mount)

    err = mount.ReadDefaultConfigFile()
    assert.NoError(t, err)

    err = mount.Mount()
    assert.NoError(t, err)

    dir := mount.CurrentDir()
    assert.NotNil(t, dir)
}
