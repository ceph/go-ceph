package cephfs_test

import (
	"fmt"
	"github.com/ceph/go-ceph/cephfs"
	"github.com/stretchr/testify/assert"
	"os"
	"syscall"
	"testing"
)

var (
	CephMountTest string = "/tmp/ceph/mds/mnt/"
)

func TestCreateMount(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)
}

func TestMountRoot(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)
}

func TestSyncFs(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)
}

func TestChangeDir(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
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

	err = mount.RemoveDir("/asdf")
	assert.NoError(t, err)
}

func TestRemoveDir(t *testing.T) {
	dirname := "one"
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	err = mount.MakeDir(dirname, 0755)
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
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)
	fmt.Printf("%#v\n", mount.IsMounted())

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)
	assert.True(t, mount.IsMounted())

	err = mount.Unmount()
	assert.NoError(t, err)
	assert.False(t, mount.IsMounted())
}

func TestReleaseMount(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.Release()
	assert.NoError(t, err)
}

func TestChmod(t *testing.T) {
	dirname := "two"
	var stats_before uint32 = 0755
	var stats_after uint32 = 0700
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	err = mount.MakeDir(dirname, stats_before)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(CephMountTest + dirname)
	assert.NoError(t, err)

	assert.Equal(t, uint32(stats.Mode().Perm()), stats_before)

	err = mount.Chmod(dirname, stats_after)
	assert.NoError(t, err)

	stats, err = os.Stat(CephMountTest + dirname)
	assert.Equal(t, uint32(stats.Mode().Perm()), stats_after)

	err = mount.RemoveDir(dirname)
	assert.NoError(t, err)

	fd, err := mount.Open("/text.txt", os.O_CREATE, stats_before)
	assert.NoError(t, err)

	err = mount.Fchmod(fd, stats_after)
	assert.NoError(t, err)

	stats, err = os.Stat(CephMountTest + "text.txt")
	assert.Equal(t, uint32(stats.Mode().Perm()), stats_after)

	err = mount.Close(fd)
	assert.NoError(t, err)

	err = mount.Unlink("/text.txt")
	assert.NoError(t, err)
}

// Not cross-platform, go's os does not specifiy Sys return type
func TestChown(t *testing.T) {
	dirname := "three"
	// dockerfile creates bob user account
	var bob uint32 = 1010
	var root uint32 = 0

	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	err = mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(CephMountTest + dirname)
	assert.NoError(t, err)

	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), root)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), root)

	err = mount.Chown(dirname, bob, bob)
	assert.NoError(t, err)

	stats, err = os.Stat(CephMountTest + dirname)
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), bob)

	err = mount.RemoveDir(dirname)
	assert.NoError(t, err)
}

func TestSetGetConf(t *testing.T) {
	value := "cephx"
	option := "auth supported"

	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.SetConf(option, value)
	assert.NoError(t, err)

	newValue, err := mount.GetConf(option)
	assert.NoError(t, err)

	assert.Equal(t, value, newValue)
}

func TestOpenClose(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	fd, err := mount.Open("/text.txt", os.O_CREATE, 0755)
	assert.NoError(t, err)

	err = mount.Close(fd)
	assert.NoError(t, err)

	err = mount.Unlink("/text.txt")
	assert.NoError(t, err)
}

func TestWriteRead(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	fd, err := mount.Open("/text.txt", os.O_CREATE|os.O_RDWR, 0755)
	assert.NoError(t, err)

	data := []byte("Ceph uniquely delivers object, block, and file storage in one unified system.")

	size, err := mount.Write(fd, data, uint64(len(data)), 0)
	assert.NoError(t, err)
	assert.Equal(t, size, len(data))

	readData, err := mount.Read(fd, uint64(len(data)), 0)
	assert.NoError(t, err)
	assert.Equal(t, data, readData)

	err = mount.Close(fd)
	assert.NoError(t, err)

	err = mount.Unlink("/text.txt")
	assert.NoError(t, err)
}

func TestListDir(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	err = mount.MakeDir("/testdir", 0755)
	assert.NoError(t, err)

	assert.NoError(t, err)
	l, err := mount.ListDir("/")
	assert.NoError(t, err)
	assert.NotNil(t, l)

	err = mount.RemoveDir("/testdir")
	assert.NoError(t, err)
}

func TestStatLink(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	fd, err := mount.Open("/text.txt", os.O_CREATE, 0755)
	assert.NoError(t, err)
	err = mount.Close(fd)
	assert.NoError(t, err)

	stat, err := mount.LStat("/text.txt")
	assert.Equal(t, stat.IsFile, true)

	err = mount.MakeDir("/testdir", 0755)
	assert.NoError(t, err)
	stat, err = mount.Stat("/testdir")
	assert.Equal(t, stat.IsDir, true)

	err = mount.Link("/text.txt", "/hardLink")
	assert.NoError(t, err)
	stat, err = mount.LStat("/hardLink")
	assert.Equal(t, stat.IsSymlink, false)

	err = mount.Symlink("/text.txt", "/link")
	assert.NoError(t, err)
	stat, err = mount.Stat("/link")
	assert.Equal(t, stat.IsSymlink, false)
	stat, err = mount.LStat("/link")
	assert.Equal(t, stat.IsSymlink, true)

	/*
	   name, err := mount.ReadLink("/link")
	   assert.NoError(t, err)
	   assert.Equal(t, "/text.txt", name)
	*/

	err = mount.Unlink("/link")
	assert.NoError(t, err)
	err = mount.Unlink("/hardLink")
	assert.NoError(t, err)
	err = mount.Unlink("/text.txt")
	assert.NoError(t, err)

	err = mount.RemoveDir("/testdir")
	assert.NoError(t, err)
}

func TestTruncate(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	fd, err := mount.Open("/text.txt", os.O_CREATE|os.O_RDWR, 0755)
	assert.NoError(t, err)

	data := []byte("Ceph uniquely delivers object, block, and file storage in one unified system.")

	size, err := mount.Write(fd, data, uint64(len(data)), 0)
	assert.NoError(t, err)
	assert.Equal(t, size, len(data))

	stat, err := mount.Stat("/text.txt")
	assert.Equal(t, int(stat.Size), len(data))

	err = mount.FTruncate(fd, 20)
	assert.NoError(t, err)
	stat, err = mount.Stat("/text.txt")
	assert.Equal(t, int(stat.Size), 20)

	err = mount.Truncate("/text.txt", 10)
	assert.NoError(t, err)
	stat, err = mount.Stat("/text.txt")
	assert.Equal(t, int(stat.Size), 10)

	err = mount.Close(fd)
	assert.NoError(t, err)

	err = mount.Unlink("/text.txt")
	assert.NoError(t, err)
}

func TestLseek(t *testing.T) {
	mount, err := cephfs.CreateMount("")
	assert.NoError(t, err)
	assert.NotNil(t, mount)

	err = mount.ReadDefaultConfigFile()
	assert.NoError(t, err)

	err = mount.Mount("/")
	assert.NoError(t, err)

	fd, err := mount.Open("/text.txt", os.O_CREATE|os.O_RDWR, 0755)
	assert.NoError(t, err)

	data := []byte("Ceph uniquely delivers object, block, and file storage in one unified system.")

	size, err := mount.Write(fd, data, uint64(len(data)), 0)
	assert.NoError(t, err)
	assert.Equal(t, size, len(data))

	err = mount.Lseek(fd, 5, os.SEEK_SET)
	assert.NoError(t, err)

	err = mount.Close(fd)
	assert.NoError(t, err)

	err = mount.Unlink("/text.txt")
	assert.NoError(t, err)
}
