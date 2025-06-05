package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChmodDir(t *testing.T) {
	dirname := "two"
	var statsBefore uint32 = 0755
	var statsAfter uint32 = 0700
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, statsBefore)
	assert.NoError(t, err)
	defer mount.RemoveDir(dirname)

	err = mount.SyncFs()
	assert.NoError(t, err)

	sx, err := mount.Statx(dirname, StatxBasicStats, 0)
	require.NoError(t, err)

	assert.Equal(t, uint32(sx.Mode&0777), statsBefore)

	err = mount.Chmod(dirname, statsAfter)
	assert.NoError(t, err)

	sx, err = mount.Statx(dirname, StatxBasicStats, 0)
	require.NoError(t, err)
	assert.Equal(t, uint32(sx.Mode&0777), statsAfter)
}

func TestChown(t *testing.T) {
	dirname := "three"
	// dockerfile creates bob user account
	var bob uint32 = 1010
	var root uint32

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)
	defer mount.RemoveDir(dirname)

	err = mount.SyncFs()
	assert.NoError(t, err)

	sx, err := mount.Statx(dirname, StatxBasicStats, 0)
	require.NoError(t, err)

	assert.Equal(t, sx.Uid, root)
	assert.Equal(t, sx.Gid, root)

	err = mount.Chown(dirname, bob, bob)
	assert.NoError(t, err)

	sx, err = mount.Statx(dirname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, sx.Uid, bob)
	assert.Equal(t, sx.Gid, bob)
}

func TestLchown(t *testing.T) {
	dirname := "four"
	var bob uint32 = 1010
	var root uint32

	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, 0755)
	assert.NoError(t, err)
	defer mount.RemoveDir(dirname)

	err = mount.SyncFs()
	assert.NoError(t, err)

	err = mount.Symlink(dirname, "symlnk")
	assert.NoError(t, err)
	defer mount.Unlink("symlnk")

	err = mount.Lchown("symlnk", bob, bob)
	sx, err := mount.Statx("symlnk", StatxBasicStats, AtSymlinkNofollow)
	assert.NoError(t, err)
	assert.Equal(t, sx.Uid, bob)
	assert.Equal(t, sx.Gid, bob)
	sx, err = mount.Statx(dirname, StatxBasicStats, AtSymlinkNofollow)
	assert.Equal(t, sx.Uid, root)
	assert.Equal(t, sx.Gid, root)
}
