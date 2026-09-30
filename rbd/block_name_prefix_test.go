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
		// Create pools until we get an id where the prefix
		// rbd_data.<metadata pool id>.<image id> exceeds
		// the fixed length used by Stat
		const statPrefixLen = 23
		dataPoolname := GetUUID()
		require.NoError(t, conn.MakePool(dataPoolname))
		defer conn.DeletePool(dataPoolname)

		// tryLongPrefix returns false if the prefix fits in Stat's field.
		tryLongPrefix := func() bool {
			mdPoolname := GetUUID()
			require.NoError(t, conn.MakePool(mdPoolname))
			defer conn.DeletePool(mdPoolname)
			mdIoctx, err := conn.OpenIOContext(mdPoolname)
			require.NoError(t, err)
			defer mdIoctx.Destroy()

			dpName := GetUUID()
			dpOptions := NewRbdImageOptions()
			defer dpOptions.Destroy()
			require.NoError(t, dpOptions.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
			require.NoError(t, dpOptions.SetString(ImageOptionDataPool, dataPoolname))
			require.NoError(t, CreateImage(mdIoctx, dpName, testImageSize, dpOptions))
			defer func() { assert.NoError(t, RemoveImage(mdIoctx, dpName)) }()

			image, err := OpenImageReadOnly(mdIoctx, dpName, NoSnapshot)
			require.NoError(t, err)
			defer func() { assert.NoError(t, image.Close()) }()

			id, err := image.GetId()
			require.NoError(t, err)
			expected := fmt.Sprintf("rbd_data.%d.%s", mdIoctx.GetPoolID(), id)

			prefix, err := image.GetBlockNamePrefix()
			require.NoError(t, err)
			require.Equal(t, expected, prefix)
			if len(expected) <= statPrefixLen {
				return false
			}

			info, err := image.Stat()
			require.NoError(t, err)
			assert.Equal(t, expected[:statPrefixLen], info.Block_name_prefix)
			return true
		}
		attempt := 1
		for ; !tryLongPrefix(); attempt++ {
			require.Less(t, attempt, 20,
				"no prefix longer than %d characters", statPrefixLen)
		}
		t.Logf("prefix longer than %d characters on attempt %d", statPrefixLen, attempt)
	})
}
