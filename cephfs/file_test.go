package cephfs

import (
	"io"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileOpen(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileOpen.txt"

	// idempotent open for read and write
	t.Run("create", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		err = f1.Close()
		assert.NoError(t, err)
		assert.NoError(t, mount.Unlink(fname))
	})

	t.Run("errorMissing", func(t *testing.T) {
		// try to open a file we know should not exist
		f2, err := mount.Open(".nope", os.O_RDONLY, 0644)
		assert.Error(t, err)
		assert.Nil(t, f2)
	})

	t.Run("existsInMount", func(t *testing.T) {
		useMount(t)

		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		err = f1.Close()
		assert.NoError(t, err)
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		s, err := os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 0, s.Size())
	})

	t.Run("idempotentClose", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		assert.NoError(t, f1.Close())
		assert.NoError(t, f1.Close()) // call close again. it should not fail
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()
	})

	t.Run("uninitializedFileClose", func(t *testing.T) {
		f := &File{}
		err := f.Close()
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})

	t.Run("invalidFdClose", func(t *testing.T) {
		f := &File{mount, 1980}
		err := f.Close()
		assert.Error(t, err)
	})

	t.Run("openInvalidMount", func(t *testing.T) {
		m := &MountInfo{}
		_, err := m.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.Error(t, err)
	})
}

func TestFileReadWrite(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileReadWrite.txt"

	t.Run("writeAndRead", func(t *testing.T) {
		// idempotent open for read and write
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		n, err := f1.Write([]byte("yello world!"))
		assert.NoError(t, err)
		assert.EqualValues(t, 12, n)
		err = f1.Close()
		assert.NoError(t, err)
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		buf := make([]byte, 1024)
		f2, err := mount.Open(fname, os.O_RDONLY, 0)
		n, err = f2.Read(buf)
		assert.NoError(t, err)
		assert.EqualValues(t, 12, n)
		assert.Equal(t, "yello world!", string(buf[:n]))
	})

	t.Run("openForWriteOnly", func(t *testing.T) {
		buf := make([]byte, 32)
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		_, err = f1.Read(buf)
		assert.Error(t, err)
	})

	t.Run("openForReadOnly", func(t *testing.T) {
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		_, err = f1.Write([]byte("yo"))
		assert.Error(t, err)
	})

	t.Run("uninitializedFile", func(t *testing.T) {
		f := &File{}
		b := []byte("testme")
		_, err := f.Write(b)
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
		_, err = f.Read(b)
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestFileReadWriteAt(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileReadWriteAt.txt"

	t.Run("writeAtAndReadAt", func(t *testing.T) {
		// idempotent open for read and write
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		n, err := f1.WriteAt([]byte("foo"), 0)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, n)
		n, err = f1.WriteAt([]byte("bar"), 6)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, n)
		// assert that negative offsets return an error
		_, err = f1.WriteAt([]byte("baz"), -10)
		assert.Error(t, err)
		err = f1.Close()
		assert.NoError(t, err)
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		buf := make([]byte, 4)
		f2, err := mount.Open(fname, os.O_RDONLY, 0)
		n, err = f2.ReadAt(buf, 0)
		assert.NoError(t, err)
		assert.EqualValues(t, 4, n)
		assert.Equal(t, "foo", string(buf[:3]))
		assert.EqualValues(t, 0, string(buf[3]))
		n, err = f2.ReadAt(buf, 6)
		assert.NoError(t, err)
		assert.EqualValues(t, 3, n)
		assert.Equal(t, "bar", string(buf[:3]))
		// assert that negative offsets return an error
		_, err = f2.ReadAt(buf, -10)
		assert.Error(t, err)
	})

	t.Run("openForWriteOnly", func(t *testing.T) {
		buf := make([]byte, 32)
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		_, err = f1.ReadAt(buf, 0)
		assert.Error(t, err)
	})

	t.Run("openForReadOnly", func(t *testing.T) {
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		_, err = f1.WriteAt([]byte("yo"), 0)
		assert.Error(t, err)
	})

	t.Run("uninitializedFile", func(t *testing.T) {
		f := &File{}
		b := []byte("testme")
		_, err := f.WriteAt(b, 0)
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
		_, err = f.ReadAt(b, 0)
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestFileInterfaces(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileInterfaces.txt"

	t.Run("ioWriter", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		var w io.Writer = f1
		_, err = w.Write([]byte("foo"))
		assert.NoError(t, err)
	})

	t.Run("ioReader", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		_, err = f1.Write([]byte("foo"))
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		var r io.Reader = f1
		buf := make([]byte, 32)
		_, err = r.Read(buf)
		assert.NoError(t, err)
		n, err := r.Read(buf)
		assert.Equal(t, 0, n)
		assert.Equal(t, io.EOF, err)
	})
}

func TestFileSeek(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileSeek.txt"

	t.Run("validSeek", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		o, err := f1.Seek(8, SeekSet)
		assert.NoError(t, err)
		assert.EqualValues(t, 8, o)

		n, err := f1.Write([]byte("flimflam"))
		assert.NoError(t, err)
		assert.EqualValues(t, 8, n)
	})

	t.Run("invalidWhence", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		o, err := f1.Seek(8, 1776)
		assert.Error(t, err)
		assert.EqualValues(t, 0, o)
	})

	t.Run("invalidSeek", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		o, err := f1.Seek(-22, SeekSet)
		assert.Error(t, err)
		assert.EqualValues(t, 0, o)
	})

	t.Run("uninitializedFile", func(t *testing.T) {
		f := &File{}
		_, err := f.Seek(0, SeekSet)
		assert.Error(t, err)
		assert.Equal(t, ErrNotConnected, err)
	})
}

func TestMixedReadReadAt(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestMixedReadReadAt.txt"

	f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	_, err = f1.Write([]byte("abc def ghi wow!"))
	assert.NoError(t, err)
	assert.NoError(t, f1.Close())
	defer func() { assert.NoError(t, mount.Unlink(fname)) }()

	buf := make([]byte, 4)
	f2, err := mount.Open(fname, os.O_RDONLY, 0)
	n, err := f2.ReadAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "abc ", string(buf))

	n, err = f2.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "abc ", string(buf))

	n, err = f2.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "def ", string(buf))

	n, err = f2.ReadAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "abc ", string(buf))

	n, err = f2.ReadAt(buf, 12)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "wow!", string(buf))

	n, err = f2.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "ghi ", string(buf))

	assert.NoError(t, f1.Close())
}

func TestFchmod(t *testing.T) {
	useMount(t)
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "file.txt"
	var statsBefore uint32 = 0755
	var statsAfter uint32 = 0700

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, statsBefore)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	err = mount.SyncFs()
	assert.NoError(t, err)

	stats, err := os.Stat(path.Join(CephMountDir, fname))
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Mode().Perm()), statsBefore)

	err = f1.Fchmod(statsAfter)
	assert.NoError(t, err)

	stats, err = os.Stat(path.Join(CephMountDir, fname))
	assert.Equal(t, uint32(stats.Mode().Perm()), statsAfter)

	// TODO use t.Run sub-tests where appropriate
	f2 := &File{}
	err = f2.Fchmod(statsAfter)
	assert.Error(t, err)
}

func TestFchown(t *testing.T) {
	useMount(t)

	fname := "file.txt"
	// dockerfile creates bob user account
	var bob uint32 = 1010
	var root uint32

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	err = mount.SyncFs()
	assert.NoError(t, err)

	stats, err := os.Stat(path.Join(CephMountDir, fname))
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), root)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), root)

	err = f1.Fchown(bob, bob)
	assert.NoError(t, err)

	stats, err = os.Stat(path.Join(CephMountDir, fname))
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), bob)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), bob)

	// TODO use t.Run sub-tests where appropriate
	f2 := &File{}
	err = f2.Fchown(bob, bob)
	assert.Error(t, err)
}

func TestFstatx(t *testing.T) {
	fname := "test_fstatx.txt"

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0600)
	assert.NoError(t, err)
	assert.NotNil(t, f)
	assert.NoError(t, f.Close())
	defer func() { assert.NoError(t, mount.Unlink(fname)) }()

	t.Run("emptyFile", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR, 0600)
		assert.NoError(t, err)
		assert.NotNil(t, f)

		st, err := f.Fstatx(StatxBasicStats, 0)
		assert.NoError(t, err)
		assert.NotNil(t, f)

		assert.Equal(t, uint16(0600), st.Mode&0600)
		assert.Equal(t, uint64(0), st.Size)
	})

	t.Run("populateFile", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0600)
		assert.NoError(t, err)
		assert.NotNil(t, f)
		defer func() { assert.NoError(t, f.Close()) }()

		_, err = f.Write([]byte("See spot run.\nSee spot jump.\n"))
		assert.NoError(t, err)

		st, err := f.Fstatx(StatxBasicStats, 0)
		assert.NoError(t, err)
		assert.NotNil(t, f)

		assert.Equal(t, uint16(0600), st.Mode&0600)
		assert.Equal(t, uint64(29), st.Size)
	})

	t.Run("closedFile", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0600)
		assert.NoError(t, err)
		assert.NotNil(t, f)
		assert.NoError(t, f.Close())

		st, err := f.Fstatx(StatxBasicStats, 0)
		assert.Error(t, err)
		assert.Nil(t, st)
	})

	t.Run("invalidFile", func(t *testing.T) {
		f := &File{}
		_, err := f.Fstatx(StatxBasicStats, 0)
		assert.Error(t, err)
	})
}

func TestFallocate(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "file1.txt"
	f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f)
	defer func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	// assert that negative values will return error.
	t.Run("NegativeOffsetLength", func(t *testing.T) {
		err = f.Fallocate(FallocNoFlag, -1, 10)
		assert.Error(t, err)

		err = f.Fallocate(FallocNoFlag, 10, -1)
		assert.Error(t, err)
	})

	// Allocate space - default case, mode == 0.
	t.Run("modeIsZero", func(t *testing.T) {
		useMount(t)
		// check file size.
		s, err := os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 0, s.Size())
		// write 10 bytes at offset 0.
		err = f.Fallocate(FallocNoFlag, 0, 10)
		assert.NoError(t, err)
		// check file size again.
		s, err = os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 10, s.Size())
	})

	// Allocate space - size increases, data remains intact.
	t.Run("increaseSize", func(t *testing.T) {
		useMount(t)
		fname := "file2.txt"
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		// write to file.
		n, err := f1.Write([]byte("Ten chars!"))
		assert.NoError(t, err)
		assert.EqualValues(t, 10, n)
		// check the file size.
		s, err := os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 10, s.Size())
		// allocate 10 more bytes from the middle.
		err = f1.Fallocate(FallocNoFlag, 5, 10)
		assert.NoError(t, err)
		// check the size, it should increase.
		s, err = os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 15, s.Size())
		// Read the contents, first ten chars remain intact.
		buf := make([]byte, 10)
		n, err = f1.ReadAt(buf, 0)
		assert.NoError(t, err)
		assert.Equal(t, "Ten chars!", string(buf[:n]))
	})

	// Allocate space - with FALLOC_FL_KEEP_SIZE.
	t.Run("allocateSpaceWithFlag", func(t *testing.T) {
		useMount(t)
		fname := "file3.txt"
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		// Write to file.
		n, err := f1.Write([]byte("tenchars!!"))
		assert.NoError(t, err)
		assert.EqualValues(t, 10, n)
		// Allocate 10 more bytes from the middle.
		err = f1.Fallocate(FallocFlKeepSize, 5, 10)
		assert.NoError(t, err)
		// Check the file size, it should not increase.
		s, err := os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 10, s.Size())
	})

	// Deallocate space - with only FALLOC_FL_PUNCH_HOLE.
	t.Run("punchHoleFlagAlone", func(t *testing.T) {
		err := f.Fallocate(FallocFlPunchHole, 0, 10)
		// Not supported.
		assert.Error(t, err)
	})

	// De-allocate space - punch holes.
	t.Run("punchActualHoles", func(t *testing.T) {
		fname := "file4.txt"
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		// Write some data.
		n, err := f1.Write([]byte("Ten chars!"))
		assert.NoError(t, err)
		assert.EqualValues(t, 10, n)
		// Read it back.
		buf := make([]byte, 10)
		n, err = f1.ReadAt(buf, 0)
		assert.NoError(t, err)
		assert.Equal(t, "Ten chars!", string(buf[:n]))
		// Punch holes.
		err = f1.Fallocate(FallocFlPunchHole|FallocFlKeepSize, 0, 5)
		assert.NoError(t, err)
		// Read again - first five chars.
		buf = make([]byte, 5)
		n, err = f1.ReadAt(buf, 0)
		assert.NoError(t, err)
		assert.Equal(t, "\x00\x00\x00\x00\x00", string(buf[:n]))
		// Read again - last five chars.
		n, err = f1.ReadAt(buf, 5)
		assert.Equal(t, "hars!", string(buf[:n]))
	})

	t.Run("checkValidate", func(t *testing.T) {
		f1 := &File{}
		err := f1.Fallocate(FallocNoFlag, 0, 10)
		assert.Error(t, err)
	})
}

func TestFlock(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	t.Run("validate", func(t *testing.T) {
		f := &File{}
		err := f.Flock(LockSH, 1010)
		assert.Error(t, err)
	})

	t.Run("validateOperation", func(t *testing.T) {
		fname := "Flockfile.txt"
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f)
		defer func() {
			assert.NoError(t, f.Close())
			assert.NoError(t, mount.Unlink(fname))
		}()
		err = f.Flock(LockSH|LockEX, 1010)
		assert.Error(t, err)
	})

	t.Run("basicLocking", func(t *testing.T) {
		const (
			anna  = 42
			bob   = 43
			chris = 44
		)
		fname1 := "Flockfile1.txt"
		f1, err := mount.Open(fname1, os.O_RDWR|os.O_CREATE, 0644)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		defer func() {
			assert.NoError(t, f1.Close())
			assert.NoError(t, mount.Unlink(fname1))
		}()
		// Lock exclusively twice.
		t.Run("exclusiveTwiceBlock", func(t *testing.T) {
			err := f1.Flock(LockEX, anna)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, f1.Flock(LockUN, anna))
			}()
			err = f1.Flock(LockEX|LockNB, bob)
			assert.Error(t, err)
		})
		t.Run("exclusiveTwiceNonBlock", func(t *testing.T) {
			err := f1.Flock(LockEX|LockNB, anna)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, f1.Flock(LockUN, anna))
			}()
			err = f1.Flock(LockEX|LockNB, bob)
			assert.Error(t, err)
		})

		// Lock shared.
		t.Run("sharedLock", func(t *testing.T) {
			err := f1.Flock(LockSH, anna)
			assert.NoError(t, err)
			err = f1.Flock(LockSH, bob)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, f1.Flock(LockUN, anna))
				assert.NoError(t, f1.Flock(LockUN, bob))
			}()
			// Now try to take exclusive lock.
			err = f1.Flock(LockEX|LockNB, chris)
			assert.Error(t, err)
		})

		// Lock shared with upgrade to exclusive.
		t.Run("sharedLockUpExclusive", func(t *testing.T) {
			err := f1.Flock(LockSH, bob)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, f1.Flock(LockUN, bob))
			}()
			err = f1.Flock(LockEX, bob)
			assert.NoError(t, err)
		})

		// Lock exclusive with downgrade to shared.
		t.Run("exclusiveLockDownShared", func(t *testing.T) {
			err := f1.Flock(LockEX, bob)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, f1.Flock(LockUN, bob))
			}()
			err = f1.Flock(LockSH, bob)
			assert.NoError(t, err)
		})
	})
}

func TestFsync(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "test_fsync.txt"
	defer mount.Unlink(fname)

	// unfortunately there's not much to assert around the the behavior of
	// fsync in these simple tests so we sort-of have to trust ceph on this :-)
	t.Run("simpleFsync", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		defer func() { assert.NoError(t, f.Close()) }()
		assert.NoError(t, err)
		_, err = f.Write([]byte("batman"))
		assert.NoError(t, err)
		err = f.Fsync(SyncAll)
		assert.NoError(t, err)
	})
	t.Run("DataOnly", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		defer func() { assert.NoError(t, f.Close()) }()
		assert.NoError(t, err)
		_, err = f.Write([]byte("superman"))
		assert.NoError(t, err)
		err = f.Fsync(SyncDataOnly)
		assert.NoError(t, err)
	})
	t.Run("invalid", func(t *testing.T) {
		f := &File{}
		err := f.Fsync(SyncAll)
		assert.Error(t, err)
	})
}

func TestSync(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "test_sync.txt"
	defer mount.Unlink(fname)

	// see fsync
	t.Run("simple", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
		defer func() { assert.NoError(t, f.Close()) }()
		assert.NoError(t, err)
		_, err = f.Write([]byte("question"))
		assert.NoError(t, err)
		err = f.Sync()
		assert.NoError(t, err)
	})
	t.Run("invalid", func(t *testing.T) {
		f := &File{}
		err := f.Sync()
		assert.Error(t, err)
	})
}

func TestFilePreadvPwritev(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "TestFilePreadvPwritev.txt"
	defer mount.Unlink(fname)

	t.Run("simple", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f.Close()) }()

		b1 := []byte("foobarbaz")
		b2 := []byte("alphabeta")
		b3 := []byte("superawseomefuntime")
		n, err := f.Pwritev([][]byte{b1, b2, b3}, 0)
		assert.NoError(t, err)
		assert.Equal(t, 37, n)

		o := [][]byte{
			make([]byte, 3),
			make([]byte, 3),
			make([]byte, 3),
			make([]byte, 3),
		}
		n, err = f.Preadv(o, 0)
		assert.NoError(t, err)
		assert.Equal(t, 12, n)
		assert.Equal(t, "foo", string(o[0]))
		assert.Equal(t, "bar", string(o[1]))
		assert.Equal(t, "baz", string(o[2]))
		assert.Equal(t, "alp", string(o[3]))
	})

	t.Run("silly", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f.Close()) }()

		b := []byte("foo")
		x := make([][]byte, 8)
		for i := range x {
			x[i] = b
		}
		n, err := f.Pwritev(x, 0)
		assert.NoError(t, err)
		assert.Equal(t, 24, n)

		for i := range x {
			x[i] = make([]byte, 6)
		}
		n, err = f.Preadv(x, 1)
		assert.NoError(t, err)
		assert.Equal(t, 23, n)
		assert.Equal(t, "oofoof", string(x[0]))
		assert.Equal(t, "oofoof", string(x[1]))
		assert.Equal(t, "oofoof", string(x[2]))
		assert.Equal(t, "oofoo\x00", string(x[3]))
	})

	t.Run("readEOF", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f.Close()) }()

		x := make([][]byte, 8)
		for i := range x {
			x[i] = make([]byte, 6)
		}
		n, err := f.Preadv(x, 16)
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, 0, n)
		assert.Equal(t, make([]byte, 6), x[0])
	})

	t.Run("shortRead", func(t *testing.T) {
		f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f.Close()) }()

		b := []byte("Antidisestablishmentarianism\n")
		n, err := f.Pwritev([][]byte{b}, 0)
		assert.NoError(t, err)
		assert.Equal(t, 29, n)

		// this is an explicit short read test.
		// some of the buffers in the vector will be left unfilled.
		x := make([][]byte, 8)
		for i := range x {
			x[i] = make([]byte, 6)
		}
		n, err = f.Preadv(x, 0)
		assert.NoError(t, err)
		assert.Equal(t, 29, n)
		assert.Equal(t, "Antidi", string(x[0]))
		assert.Equal(t, "sestab", string(x[1]))
		assert.Equal(t, "lishme", string(x[2]))
		assert.Equal(t, "ntaria", string(x[3]))
		assert.Equal(t, "nism\n\x00", string(x[4]))
		assert.Equal(t, make([]byte, 6), x[5])
		assert.Equal(t, make([]byte, 6), x[6])
		assert.Equal(t, make([]byte, 6), x[7])
	})

	t.Run("openForWriteOnly", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()

		x := make([][]byte, 8)
		for i := range x {
			x[i] = make([]byte, 6)
		}
		_, err = f1.Preadv(x, 0)
		assert.Error(t, err)
	})

	t.Run("openForReadOnly", func(t *testing.T) {
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0644)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()

		x := make([][]byte, 8)
		for i := range x {
			x[i] = []byte("robble")
		}
		_, err = f1.Pwritev(x, 0)
		assert.Error(t, err)
	})

	t.Run("writeInvalidFile", func(t *testing.T) {
		f := &File{}
		_, err := f.Pwritev([][]byte{}, 0)
		assert.Error(t, err)
	})

	t.Run("readInvalidFile", func(t *testing.T) {
		f := &File{}
		_, err := f.Preadv([][]byte{}, 0)
		assert.Error(t, err)
	})
}

func TestFileTruncate(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "TestFileTruncate.txt"
	defer mount.Unlink(fname)

	t.Run("invalidSize", func(t *testing.T) {
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, f1.Close())
		}()

		err = f1.Truncate(-1)
		assert.Error(t, err)

		st, err := f1.Fstatx(StatxBasicStats, 0)
		if assert.NoError(t, err) {
			assert.EqualValues(t, 0, st.Size)
		}
	})

	t.Run("closedFile", func(t *testing.T) {
		t.Skip("test fails because of a bug(?) in ceph")
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f2, err := mount.Open(fname, os.O_RDONLY, 0644)
		assert.NoError(t, err)
		assert.NoError(t, f2.Close())

		err = f2.Truncate(1024)
		assert.Error(t, err)

		// I wanted to do the stat check here too but it is a pain to implement
		// because we close the file.
		// The original version of this test, using a read-only file, failed
		// due to a bug in libcephfs (see Truncate doc comment).
	})

	t.Run("invalidFile", func(t *testing.T) {
		f := &File{}
		err := f.Truncate(0)
		assert.Error(t, err)
	})
}
