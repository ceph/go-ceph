//go:build ceph_preview
// +build ceph_preview

package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDirs(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	dir1 := "/base/sub/way"
	err := mount.MakeDirs(dir1, 0755)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, mount.RemoveDir("/base/sub/way"))
		assert.NoError(t, mount.RemoveDir("/base/sub"))
		assert.NoError(t, mount.RemoveDir("/base"))
	}()

	dir, err := mount.OpenDir(dir1)
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	err = dir.Close()
	assert.NoError(t, err)
}
