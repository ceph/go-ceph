//go:build !nautilus
// +build !nautilus

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSparsifyWithProgress(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	require.NoError(t, err)
	defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

	t.Run("valid", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, img.Close()) }()

		cc := 0
		cb := func(offset, total uint64, v interface{}) int {
			cc++
			val := v.(int)
			assert.Equal(t, 0, val)
			assert.Equal(t, uint64(1), total)
			return 0
		}

		err = img.SparsifyWithProgress(4096, cb, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, cc, 1)
	})

	t.Run("negativeReturnValue", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, img.Close()) }()

		cc := 0
		cb := func(offset, total uint64, v interface{}) int {
			cc++
			val := v.(int)
			assert.Equal(t, 0, val)
			assert.Equal(t, uint64(1), total)
			return -1
		}

		err = img.SparsifyWithProgress(4096, cb, 0)
		assert.Error(t, err)
	})

	t.Run("closedImage", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())

		cc := 0
		cb := func(offset, total uint64, v interface{}) int {
			cc++
			val := v.(int)
			assert.Equal(t, 0, val)
			assert.Equal(t, uint64(1), total)
			return 0
		}

		err = img.SparsifyWithProgress(4096, cb, 0)
		assert.Error(t, err)
	})

	t.Run("invalidValue", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, img.Close()) }()

		err = img.SparsifyWithProgress(4096, nil, nil)
		assert.Error(t, err)
	})
}
