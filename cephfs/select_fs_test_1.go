package cephfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	altFSName = ""
)

func init() {
	altFSName = os.Getenv("GO_CEPH_TEST_ALT_FS_NAME")
}

func TestSelectFS(t *testing.T) {
	if altFSName == "" {
		t.Skip("no alternative fs provided")
	}

	t.Run("selectFilesystem", func(t *testing.T) {
		mount, err := CreateMount()
		assert.NoError(t, err)
		assert.NotNil(t, mount)

		err = mount.SelectFilesystem(altFSName)
		assert.NoError(t, err)

		assert.NoError(t, mount.Release())
	})

	t.Run("selectFilesystemError", func(t *testing.T) {
		mount := fsConnect(t)
		defer fsDisconnect(t, mount)

		// already mounted - this should return an error
		err := mount.SelectFilesystem(altFSName)
		assert.Error(t, err)
	})

	t.Run("invalidName", func(t *testing.T) {
		mount := fsConnect(t)
		defer func() {
			assert.NoError(t, mount.Release())
		}()

		assert.NoError(t, mount.Unmount())
		// this call will not fail because the name isn't used until
		// the file system is "mounted"
		err := mount.SelectFilesystem("a.bunch-of~nonsense")
		assert.NoError(t, err)

		// this call is the one that fails because of the invalid name
		assert.Error(t, mount.Mount())
	})

	t.Run("swapFS", func(t *testing.T) {
		mount := fsConnect(t)
		defer fsDisconnect(t, mount)

		// create a file on the first file system
		fname := "first_fs.txt"
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())
		_, err = mount.Statx(fname, StatxBasicStats, 0)
		assert.NoError(t, err)

		// swap file systems - unmount, select fs, and then mount again
		assert.NoError(t, mount.Unmount())
		err = mount.SelectFilesystem(altFSName)
		assert.NoError(t, err)
		assert.NoError(t, mount.Mount())

		// now we're on a new fs. stat'ing the file should fail, as the file is
		// on the other file system
		_, err = mount.Statx(fname, StatxBasicStats, 0)
		assert.Error(t, err)

		// verify that other operations on the 2nd fs work the same
		fname2 := "second_fs.txt"
		f2, err := mount.Open(fname2, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f2.Close())
		_, err = mount.Statx(fname2, StatxBasicStats, 0)
		assert.NoError(t, err)
		assert.NoError(t, mount.Unlink(fname2))

		// swap back to the first file system
		assert.NoError(t, mount.Unmount())
		// first fs is always called cephfs for go-ceph tests
		err = mount.SelectFilesystem("cephfs")
		assert.NoError(t, err)
		assert.NoError(t, mount.Mount())

		// we're back on the first fs. see that the file still exists and then
		// clean it up
		_, err = mount.Statx(fname, StatxBasicStats, 0)
		assert.NoError(t, err)
		assert.NoError(t, mount.Unlink(fname))
	})
}
