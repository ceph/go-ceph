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
