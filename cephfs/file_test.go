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
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		assert.NotNil(t, f1)
		err = f1.Close()
		assert.NoError(t, err)
		assert.NoError(t, mount.Unlink(fname))
	})

	t.Run("errorMissing", func(t *testing.T) {
		// try to open a file we know should not exist
		f2, err := mount.Open(".nope", os.O_RDONLY, 0666)
		assert.Error(t, err)
		assert.Nil(t, f2)
	})

	t.Run("existsInMount", func(t *testing.T) {
		useMount(t)

		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
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
		f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
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
		_, err := m.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		assert.Error(t, err)
	})
}

func TestFileReadWrite(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileReadWrite.txt"

	t.Run("writeAndRead", func(t *testing.T) {
		// idempotent open for read and write
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		_, err = f1.Read(buf)
		assert.Error(t, err)
	})

	t.Run("openForReadOnly", func(t *testing.T) {
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0666)
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
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		_, err = f1.ReadAt(buf, 0)
		assert.Error(t, err)
	})

	t.Run("openForReadOnly", func(t *testing.T) {
		// "touch" the file
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0666)
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
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		var w io.Writer = f1
		_, err = w.Write([]byte("foo"))
		assert.NoError(t, err)
	})

	t.Run("ioReader", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		_, err = f1.Write([]byte("foo"))
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0666)
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
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()
		defer func() { assert.NoError(t, mount.Unlink(fname)) }()

		o, err := f1.Seek(8, 1776)
		assert.Error(t, err)
		assert.EqualValues(t, 0, o)
	})

	t.Run("invalidSeek", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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

	f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0666)
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
}
