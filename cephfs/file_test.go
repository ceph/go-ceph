package cephfs

import (
	"io"
	"os"
	"path"
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
		// TODO: clean up file
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

		s, err := os.Stat(path.Join(CephMountDir, fname))
		assert.NoError(t, err)
		assert.EqualValues(t, 0, s.Size())
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
		err = f1.Close()
		assert.NoError(t, err)

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
}

func TestFileInterfaces(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestFileInterfaces.txt"

	t.Run("ioWriter", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()

		var w io.Writer = f1
		_, err = w.Write([]byte("foo"))
		assert.NoError(t, err)
	})

	t.Run("ioReader", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		assert.NoError(t, f1.Close())

		f1, err = mount.Open(fname, os.O_RDONLY, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()

		var r io.Reader = f1
		buf := make([]byte, 32)
		_, err = r.Read(buf)
		assert.NoError(t, err)
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

		o, err := f1.Seek(8, 1776)
		assert.Error(t, err)
		assert.EqualValues(t, 0, o)
	})

	t.Run("invalidSeek", func(t *testing.T) {
		f1, err := mount.Open(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, f1.Close()) }()

		o, err := f1.Seek(-22, SeekSet)
		assert.Error(t, err)
		assert.EqualValues(t, 0, o)
	})
}
