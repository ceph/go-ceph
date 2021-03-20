package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChmodDir(t *testing.T) {
	dirname := "two"
	var stats_before uint32 = 0755
	var stats_after uint32 = 0700
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	err := mount.MakeDir(dirname, stats_before)
	assert.NoError(t, err)
	defer mount.RemoveDir(dirname)

	err = mount.SyncFs()
	assert.NoError(t, err)

	sx, err := mount.Statx(dirname, StatxBasicStats, 0)
	require.NoError(t, err)

	assert.Equal(t, uint32(sx.Mode&0777), stats_before)

	err = mount.Chmod(dirname, stats_after)
	assert.NoError(t, err)

	sx, err = mount.Statx(dirname, StatxBasicStats, 0)
	require.NoError(t, err)
	assert.Equal(t, uint32(sx.Mode&0777), stats_after)
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

	assert.Equal(t, uint32(sx.Uid), root)
	assert.Equal(t, uint32(sx.Gid), root)

	err = mount.Chown(dirname, bob, bob)
	assert.NoError(t, err)

	sx, err = mount.Statx(dirname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, uint32(sx.Uid), bob)
	assert.Equal(t, uint32(sx.Gid), bob)
}
