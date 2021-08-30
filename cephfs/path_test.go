package cephfs

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeDir(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

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

	err = mount.ChangeDir("/")
	assert.NoError(t, err)
	err = mount.RemoveDir("/asdf")
	assert.NoError(t, err)
}

func TestRemoveDir(t *testing.T) {
	dirname := "one"
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)

	err = mount.SyncFs()
	assert.NoError(t, err)

	// Stat the location to verify dirname currently exists
	_, err = mount.Statx(dirname, StatxBasicStats, 0)
	assert.NoError(t, err)

	err = mount.RemoveDir(dirname)
	assert.NoError(t, err)

	_, err = mount.Statx(dirname, StatxBasicStats, 0)
	assert.Equal(t, err, errNoEntry)
}

func TestLink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	t.Run("rootDirOperations", func(t *testing.T) {
		// Root dir, both as source and destination.
		err := mount.Link("/", "/")
		// Error directory operations are not allowed.
		assert.Error(t, err)

		dir1 := "myDir1"
		assert.NoError(t, mount.MakeDir(dir1, 0755))
		defer func() {
			assert.NoError(t, mount.RemoveDir(dir1))
		}()

		// Creating link for a directory.
		err = mount.Link(dir1, "/")
		// Error, directory operations not allowed.
		assert.Error(t, err)
	})

	// Non-root directory operations.
	fname := "testFile.txt"
	dir2 := "myDir2"
	assert.NoError(t, mount.MakeDir(dir2, 0755))
	defer func() {
		assert.NoError(t, mount.RemoveDir(dir2))
	}()

	t.Run("dirAsSource", func(t *testing.T) {
		err := mount.Link(dir2, fname)
		// Error, directory operations not allowed.
		assert.Error(t, err)
	})

	t.Run("dirAsDestination", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE, 0666)
		assert.NotNil(t, f1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		err = mount.Link(fname, dir2)
		// Error, destination exists.
		assert.Error(t, err)
	})

	// File operations.
	t.Run("sourceDoesNotExist", func(t *testing.T) {
		fname := "notExist.txt"
		err := mount.Link(fname, "hardlnk")
		// Error, file does not exist.
		assert.Error(t, err)
	})

	t.Run("sourceExistsSuccess", func(t *testing.T) {
		fname1 := "TestFile1.txt"
		f1, err := mount.Open(fname1, os.O_WRONLY|os.O_CREATE, 0666)
		assert.NotNil(t, f1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname1))
		}()
		err = mount.Link(fname1, "hardlnk")
		defer func() { assert.NoError(t, mount.Unlink("hardlnk")) }()
		// No error, normal link operation.
		assert.NoError(t, err)
		// Verify that link got created.
		_, err = mount.Statx("hardlnk", StatxBasicStats, 0)
		assert.NoError(t, err)
	})

	t.Run("destExistsError", func(t *testing.T) {
		// Create hard link when destination exists.
		fname2 := "TestFile2.txt"
		fname3 := "TestFile3.txt"
		f2, err := mount.Open(fname2, os.O_WRONLY|os.O_CREATE, 0666)
		assert.NotNil(t, f2)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f2.Close())
			assert.NoError(t, mount.Unlink(fname2))
		}()
		f3, err := mount.Open(fname3, os.O_WRONLY|os.O_CREATE, 0666)
		assert.NotNil(t, f3)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f3.Close())
			assert.NoError(t, mount.Unlink(fname3))
		}()
		err = mount.Link(fname2, fname3)
		// Error, destination already exists.
		assert.Error(t, err)
	})
}

func TestUnlink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	t.Run("fileUnlink", func(t *testing.T) {
		fname := "TestFile.txt"
		err := mount.Unlink(fname)
		// Error, file does not exist.
		assert.Error(t, err)

		f, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE, 0666)
		assert.NotNil(t, f)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		assert.NoError(t, mount.Link(fname, "hardlnk"))

		err = mount.Unlink("hardlnk")
		// No Error, link will be removed.
		assert.NoError(t, err)
	})

	t.Run("dirUnlink", func(t *testing.T) {
		dirname := "/a"
		err := mount.MakeDir(dirname, 0755)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, mount.RemoveDir(dirname))
		}()

		err = mount.Unlink(dirname)
		// Error, not permitted on directory.
		assert.Error(t, err)
	})
}

func TestSymlink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	// File operations.
	t.Run("sourceDoesNotExistSuccess", func(t *testing.T) {
		fname1 := "TestFile1.txt"
		err := mount.Symlink(fname1, "Symlnk1")
		// No Error, symlink works even if source file doesn't exist.
		assert.NoError(t, err)
		_, err = mount.Statx("Symlnk1", StatxBasicStats, 0)
		// Error, source is not there.
		assert.Error(t, err)

		_, err = mount.Statx(fname1, StatxBasicStats, 0)
		// Error, source file is still not there.
		assert.Error(t, err)
	})

	t.Run("symlinkExistsError", func(t *testing.T) {
		fname1 := "TestFile1.txt"
		f1, err := mount.Open(fname1, os.O_RDWR|os.O_CREATE, 0666)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname1))
		}()
		err = mount.Symlink(fname1, "Symlnk1")
		// Error, Symlink1 exists.
		assert.Error(t, err)
		defer func() {
			assert.NoError(t, mount.Unlink("Symlnk1"))
		}()
	})

	t.Run("sourceExistsSuccess", func(t *testing.T) {
		fname2 := "TestFile2.txt"
		f2, err := mount.Open(fname2, os.O_RDWR|os.O_CREATE, 0666)
		assert.NoError(t, err)
		assert.NotNil(t, f2)
		defer func() {
			assert.NoError(t, f2.Close())
			assert.NoError(t, mount.Unlink(fname2))
		}()
		err = mount.Symlink(fname2, "Symlnk2")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, mount.Unlink("Symlnk2"))
		}()
		_, err = mount.Statx("Symlnk2", StatxBasicStats, 0)
		assert.NoError(t, err)
	})

	// Directory operations.
	t.Run("rootDirOps", func(t *testing.T) {
		err := mount.Symlink("/", "/")
		assert.Error(t, err)

		err = mount.Symlink("/", "someDir")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, mount.Unlink("someDir"))
		}()

		err = mount.Symlink("someFile", "/")
		// Error, permission denied.
		assert.Error(t, err)
	})

	t.Run("nonRootDir", func(t *testing.T) {
		// 1. Create a directory.
		// 2. Create a symlink to that directory.
		// 3. Create a file inside symlink.
		// 4. Ensure that it is not a directory.
		dirname := "mydir"
		err := mount.MakeDir(dirname, 0755)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, mount.RemoveDir(dirname))
		}()

		err = mount.Symlink(dirname, "symlnk")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, mount.Unlink("symlnk"))
		}()

		fname := "symlnk/file"
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0666)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		sx, err := mount.Statx("symlnk/file", StatxBasicStats, 0)
		assert.NoError(t, err)
		assert.NotEqual(t, sx.Mode&syscall.S_IFMT, uint16(syscall.S_IFDIR))
	})
}

func TestReadlink(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	t.Run("regularFile", func(t *testing.T) {
		fname := "file1.txt"
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0666)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()

		buf, err := mount.Readlink(fname)
		// Error, given path is not symbolic link.
		assert.Error(t, err)
		assert.Equal(t, buf, "")
	})

	t.Run("symLink", func(t *testing.T) {
		path1 := "path1"
		path2 := "path2"
		assert.NoError(t, mount.Symlink(path1, path2))
		defer func() {
			assert.NoError(t, mount.Unlink(path2))
		}()
		buf, err := mount.Readlink(path2)
		assert.NoError(t, err)
		assert.Equal(t, buf, path1)
	})

	t.Run("hardLink", func(t *testing.T) {
		path3 := "path3"
		path4 := "path4"
		p, err := mount.Open(path3, os.O_RDWR|os.O_CREATE, 0666)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		defer func() {
			assert.NoError(t, p.Close())
			assert.NoError(t, mount.Unlink(path3))
		}()

		assert.NoError(t, mount.Link(path3, path4))
		defer func() {
			assert.NoError(t, mount.Unlink(path4))
		}()
		buf, err := mount.Readlink(path4)
		// Error, path4 is not symbolic link.
		assert.Error(t, err)
		assert.Equal(t, buf, "")
	})
}

func TestStatx(t *testing.T) {
	t.Run("statPath", func(t *testing.T) {
		mount := fsConnect(t)
		defer fsDisconnect(t, mount)

		dirname := "statme"
		assert.NoError(t, mount.MakeDir(dirname, 0755))

		st, err := mount.Statx(dirname, StatxBasicStats, 0)
		assert.NoError(t, err)
		assert.NotNil(t, st)
		assert.Equal(t, uint16(0755), st.Mode&0777)

		assert.NoError(t, mount.RemoveDir(dirname))

		st, err = mount.Statx(dirname, StatxBasicStats, 0)
		assert.Error(t, err)
		assert.Nil(t, st)
		assert.Equal(t, errNoEntry, err)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		_, err := m.Statx("junk", StatxBasicStats, 0)
		assert.Error(t, err)
	})
}

func TestRename(t *testing.T) {
	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		err := m.Rename("foo", "bar")
		assert.Error(t, err)
	})

	t.Run("renameDir", func(t *testing.T) {
		mount := fsConnect(t)
		defer fsDisconnect(t, mount)

		n1 := "new_amsterdam"
		n2 := "new_york"
		assert.NoError(t, mount.MakeDir(n1, 0755))

		err := mount.Rename(n1, n2)
		assert.NoError(t, err)

		assert.NoError(t, mount.RemoveDir(n2))
	})
}

func TestTruncate(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "TestTruncate.txt"
	defer mount.Unlink(fname)

	// "touch" the file
	f, err := mount.Open(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NoError(t, f.Close())

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		err := m.Truncate(fname, 0)
		assert.Error(t, err)
	})

	t.Run("invalidSize", func(t *testing.T) {
		err := mount.Truncate(fname, -1)
		assert.Error(t, err)
	})

	t.Run("invalidPath", func(t *testing.T) {
		err := mount.Truncate(".Non~Existant~", 0)
		assert.Error(t, err)
	})

	t.Run("valid", func(t *testing.T) {
		err := mount.Truncate(fname, 1024)
		assert.NoError(t, err)

		st, err := mount.Statx(fname, StatxBasicStats, 0)
		if assert.NoError(t, err) {
			assert.NotNil(t, st)
			assert.EqualValues(t, 1024, st.Size)
		}

		err = mount.Truncate(fname, 0)
		assert.NoError(t, err)

		st, err = mount.Statx(fname, StatxBasicStats, 0)
		if assert.NoError(t, err) {
			assert.NotNil(t, st)
			assert.EqualValues(t, 0, st.Size)
		}
	})
}
