//go:build ceph_preview

package rbd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBlockNamePrefix(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	require.NoError(t, conn.MakePool(poolname))
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()
	options := NewRbdImageOptions()
	defer options.Destroy()
	require.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	require.NoError(t, CreateImage(ioctx, name, testImageSize, options))
	defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

	t.Run("closedImage", func(t *testing.T) {
		image := GetImage(ioctx, name)
		_, err := image.GetBlockNamePrefix()
		assert.Equal(t, ErrImageNotOpen, err)
	})

	t.Run("matchesId", func(t *testing.T) {
		image, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		defer func() { assert.NoError(t, image.Close()) }()

		id, err := image.GetId()
		require.NoError(t, err)

		prefix, err := image.GetBlockNamePrefix()
		assert.NoError(t, err)
		assert.Equal(t, "rbd_data."+id, prefix)
	})

	t.Run("dataPool", func(t *testing.T) {
		// An image with a separate data pool gets the id of its metadata
		// pool embedded in the prefix. Only then can the prefix outgrow the
		// fixed-size field used by Stat.
		dataPoolname := GetUUID()
		require.NoError(t, conn.MakePool(dataPoolname))
		defer conn.DeletePool(dataPoolname)

		dpName := GetUUID()
		dpOptions := NewRbdImageOptions()
		defer dpOptions.Destroy()
		require.NoError(t, dpOptions.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
		require.NoError(t, dpOptions.SetString(ImageOptionDataPool, dataPoolname))
		require.NoError(t, CreateImage(ioctx, dpName, testImageSize, dpOptions))
		defer func() { assert.NoError(t, RemoveImage(ioctx, dpName)) }()

		image, err := OpenImageReadOnly(ioctx, dpName, NoSnapshot)
		require.NoError(t, err)
		defer func() { assert.NoError(t, image.Close()) }()

		id, err := image.GetId()
		require.NoError(t, err)
		poolID, err := conn.GetPoolByName(poolname)
		require.NoError(t, err)
		expected := fmt.Sprintf("rbd_data.%d.%s", poolID, id)

		prefix, err := image.GetBlockNamePrefix()
		assert.NoError(t, err)
		assert.Equal(t, expected, prefix)

		// Stat truncates to RBD_MAX_BLOCK_NAME_SIZE - 1 characters
		info, err := image.Stat()
		require.NoError(t, err)
		assert.Equal(t, expected[:min(len(expected), 23)], info.Block_name_prefix)
	})
}
