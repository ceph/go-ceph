//go:build !luminous
// +build !luminous

package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetMountPerms(t *testing.T) {
	mount, err := CreateMount()
	require.NoError(t, err)
	require.NotNil(t, mount)
	defer func() { assert.NoError(t, mount.Release()) }()

	err = mount.ReadDefaultConfigFile()
	require.NoError(t, err)

	err = mount.Init()
	assert.NoError(t, err)

	uperm := NewUserPerm(0, 500, []int{0, 500, 501})
	err = mount.SetMountPerms(uperm)
	assert.NoError(t, err)

	err = mount.Mount()
	assert.NoError(t, err)
	defer func() { assert.NoError(t, mount.Unmount()) }()

	t.Run("checkStat", func(t *testing.T) {
		dirname := "/check-mount-perms"
		err := mount.MakeDir(dirname, 0755)
		assert.NoError(t, err)
		defer mount.RemoveDir(dirname)
		sx, err := mount.Statx(dirname, StatxBasicStats, 0)
		require.NoError(t, err)
		assert.EqualValues(t, sx.Gid, 500)
	})
}
