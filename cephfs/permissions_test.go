package cephfs

import (
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChmodDir(t *testing.T) {
	useMount(t)

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

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(path.Join(CephMountDir, dirname))
	require.NoError(t, err)

	assert.Equal(t, uint32(stats.Mode().Perm()), stats_before)

	err = mount.Chmod(dirname, stats_after)
	assert.NoError(t, err)

	stats, err = os.Stat(path.Join(CephMountDir, dirname))
	assert.Equal(t, uint32(stats.Mode().Perm()), stats_after)
}

// Not cross-platform, go's os does not specifiy Sys return type
func TestChown(t *testing.T) {
	useMount(t)

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

	// os.Stat the actual mounted location to verify Makedir/RemoveDir
	stats, err := os.Stat(path.Join(CephMountDir, dirname))
	require.NoError(t, err)

	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), root)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), root)

	err = mount.Chown(dirname, bob, bob)
	assert.NoError(t, err)

	stats, err = os.Stat(path.Join(CephMountDir, dirname))
	assert.NoError(t, err)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Uid), bob)
	assert.Equal(t, uint32(stats.Sys().(*syscall.Stat_t).Gid), bob)
}
