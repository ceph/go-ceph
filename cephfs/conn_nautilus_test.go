// +build !luminous,!mimic

package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFsCid(t *testing.T) {
	t.Run("unmounted", func(t *testing.T) {
		mount, err := CreateMount()
		require.NoError(t, err)
		require.NotNil(t, mount)

		err = mount.ReadDefaultConfigFile()
		require.NoError(t, err)

		cid, err := mount.GetFsCid()
		assert.Error(t, err)
		assert.Equal(t, cid, int64(0))
	})
	t.Run("mounted", func(t *testing.T) {
		mount := fsConnect(t)
		defer fsDisconnect(t, mount)

		cid, err := mount.GetFsCid()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, cid, int64(0))
	})
}
