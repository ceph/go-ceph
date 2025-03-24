//go:build ceph_preview

package cephfs

import (
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// "github.com/ceph/go-ceph/common/log"
)

func TestFSCompat(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	// set up a few dirs
	err := mount.MakeDir("fst_foo", 0755)
	require.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir("fst_foo")) }()
	err = mount.MakeDir("fst_bar", 0755)
	require.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir("fst_bar")) }()
	err = mount.MakeDir("fst_bar/fst_baz", 0755)
	require.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir("fst_bar/fst_baz")) }()

	// set up a few files
	writeFile(t, mount, "wibble.txt", []byte("nothing to see here"))
	defer func() { assert.NoError(t, mount.Unlink("wibble.txt")) }()
	writeFile(t, mount, "fst_bar/nuffin.txt", []byte(""))
	defer func() { assert.NoError(t, mount.Unlink("fst_bar/nuffin.txt")) }()
	writeFile(t, mount, "fst_bar/fst_baz/super.txt", []byte("this is my favorite file"))
	defer func() { assert.NoError(t, mount.Unlink("fst_bar/fst_baz/super.txt")) }()
	writeFile(t, mount, "boop.txt", []byte("abcdefg"))
	defer func() { assert.NoError(t, mount.Unlink("boop.txt")) }()

	// uncomment for detailed debug level logging
	// log.SetDebugf(t.Logf)

	t.Run("testFS", func(t *testing.T) {
		w := Wrap(mount)
		if err := fstest.TestFS(w, "wibble.txt", "fst_bar/nuffin.txt", "fst_bar/fst_baz/super.txt", "boop.txt"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("walkDir", func(t *testing.T) {
		w := Wrap(mount)
		dirs := []string{}
		files := []string{}
		fs.WalkDir(w, ".", func(path string, d fs.DirEntry, err error) error {
			assert.NoError(t, err)
			if d.IsDir() {
				dirs = append(dirs, path)
			} else {
				files = append(files, path)
			}
			return nil
		})
		assert.Contains(t, dirs, ".")
		assert.Contains(t, dirs, "fst_foo")
		assert.Contains(t, dirs, "fst_bar")
		assert.Contains(t, dirs, "fst_bar/fst_baz")
		assert.Contains(t, files, "wibble.txt")
		assert.Contains(t, files, "boop.txt")
		assert.Contains(t, files, "fst_bar/nuffin.txt")
		assert.Contains(t, files, "fst_bar/fst_baz/super.txt")
	})
}

func writeFile(t *testing.T, mount *MountInfo, name string, data []byte) {
	f, err := mount.Open(name, os.O_WRONLY|os.O_CREATE, 0600)
	require.NoError(t, err)
	defer func() { assert.NoError(t, f.Close()) }()
	_, err = f.Write(data)
	require.NoError(t, err)
}
