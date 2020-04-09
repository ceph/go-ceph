package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenCloseDir(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	dir1 := "/base"
	err := mount.MakeDir(dir1, 0755)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir(dir1)) }()

	dir2 := dir1 + "/a"
	err = mount.MakeDir(dir2, 0755)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir(dir2)) }()

	dir, err := mount.OpenDir(dir1)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	err = dir.Close()
	assert.NoError(t, err)

	dir, err = mount.OpenDir(dir2)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	err = dir.Close()
	assert.NoError(t, err)

	dir, err = mount.OpenDir("/no.such.dir")
	assert.Error(t, err)
	assert.Nil(t, dir)
}

func TestReadDir(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	dir1 := "/base"
	err := mount.MakeDir(dir1, 0755)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir(dir1)) }()

	subdirs := []string{"a", "bb", "ccc", "dddd"}
	for _, s := range subdirs {
		spath := dir1 + "/" + s
		err = mount.MakeDir(spath, 0755)
		assert.NoError(t, err)
		defer func(d string) {
			assert.NoError(t, mount.RemoveDir(d))
		}(spath)
	}

	t.Run("root", func(t *testing.T) {
		dir, err := mount.OpenDir("/")
		assert.NoError(t, err)
		assert.NotNil(t, dir)
		defer func() { assert.NoError(t, dir.Close()) }()

		found := []string{}
		for {
			entry, err := dir.ReadDir()
			assert.NoError(t, err)
			if entry == nil {
				break
			}
			assert.NotEqual(t, Inode(0), entry.Inode())
			assert.NotEqual(t, "", entry.Name())
			found = append(found, entry.Name())
		}
		assert.Contains(t, found, "base")
	})
	t.Run("dir1", func(t *testing.T) {
		dir, err := mount.OpenDir(dir1)
		assert.NoError(t, err)
		assert.NotNil(t, dir)
		defer func() { assert.NoError(t, dir.Close()) }()

		found := []string{}
		for {
			entry, err := dir.ReadDir()
			assert.NoError(t, err)
			if entry == nil {
				break
			}
			assert.NotEqual(t, Inode(0), entry.Inode())
			assert.NotEqual(t, "", entry.Name())
			found = append(found, entry.Name())
		}
		assert.Subset(t, found, subdirs)
	})
}

func TestDirectoryList(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	dir1 := "/base"
	err := mount.MakeDir(dir1, 0755)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.RemoveDir(dir1)) }()

	subdirs := []string{"a", "bb", "ccc", "dddd"}
	for _, s := range subdirs {
		spath := dir1 + "/" + s
		err = mount.MakeDir(spath, 0755)
		assert.NoError(t, err)
		defer func(d string) {
			assert.NoError(t, mount.RemoveDir(d))
		}(spath)
	}

	t.Run("root", func(t *testing.T) {
		dir, err := mount.OpenDir("/")
		assert.NoError(t, err)
		assert.NotNil(t, dir)
		defer func() { assert.NoError(t, dir.Close()) }()

		entries, err := dir.list()
		assert.NoError(t, err)
		assert.Greater(t, len(entries), 1)
		found := entries.names()
		assert.Contains(t, found, "base")
	})
	t.Run("dir1", func(t *testing.T) {
		dir, err := mount.OpenDir(dir1)
		assert.NoError(t, err)
		assert.NotNil(t, dir)
		defer func() { assert.NoError(t, dir.Close()) }()

		entries, err := dir.list()
		assert.NoError(t, err)
		assert.Greater(t, len(entries), 1)
		found := entries.names()
		assert.Subset(t, found, subdirs)
	})
	t.Run("dir1Twice", func(t *testing.T) {
		dir, err := mount.OpenDir(dir1)
		assert.NoError(t, err)
		assert.NotNil(t, dir)
		defer func() { assert.NoError(t, dir.Close()) }()

		entries, err := dir.list()
		assert.NoError(t, err)
		assert.Greater(t, len(entries), 1)
		found := entries.names()
		assert.Subset(t, found, subdirs)

		// verify that calling list gives a complete list
		// even after being used for the same directory already
		entries, err = dir.list()
		assert.NoError(t, err)
		assert.Greater(t, len(entries), 1)
		found = entries.names()
		assert.Subset(t, found, subdirs)
	})
}
