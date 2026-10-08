//go:build ceph_preview

package rbd

import (
	"testing"

	"github.com/ceph/go-ceph/rados"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeObjectMapImage(t *testing.T, enableObjectMap bool) (*Image, *rados.IOContext, string) {
	t.Helper()

	conn := radosConnect(t)
	t.Cleanup(conn.Shutdown)

	poolName := GetUUID()
	require.NoError(t, conn.MakePool(poolName))
	t.Cleanup(func() { require.NoError(t, conn.DeletePool(poolName)) })

	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)
	t.Cleanup(ioctx.Destroy)

	name := GetUUID()
	size := uint64(3) << testImageOrder
	_, err = Create(ioctx, name, size, testImageOrder, FeatureExclusiveLock)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, RemoveImage(ioctx, name)) })

	image, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)
	t.Cleanup(func() {
		if image.image != nil {
			require.NoError(t, image.Close())
		}
	})

	if enableObjectMap {
		require.NoError(t, image.UpdateFeatures(FeatureObjectMap|FeatureFastDiff, true))
	}

	return image, ioctx, name
}

func TestGetFlags(t *testing.T) {
	invalidMask := FlagObjectMapInvalid | FlagFastDiffInvalid

	t.Run("constants", func(t *testing.T) {
		assert.Equal(t, uint64(1), FlagObjectMapInvalid)
		assert.Equal(t, uint64(2), FlagFastDiffInvalid)
		assert.Zero(t, FlagObjectMapInvalid&FlagFastDiffInvalid)
	})

	t.Run("unopenedImage", func(t *testing.T) {
		image := GetImage(nil, "nonexistent")
		flags, err := image.GetFlags()
		assert.ErrorIs(t, err, ErrImageNotOpen)
		assert.Zero(t, flags)
	})

	image, ioctx, name := makeObjectMapImage(t, false)

	t.Run("featuresDisabled", func(t *testing.T) {
		features, err := image.GetFeatures()
		require.NoError(t, err)
		require.Zero(t, features&(FeatureObjectMap|FeatureFastDiff))

		flags, err := image.GetFlags()
		require.NoError(t, err)
		assert.Zero(t, flags&invalidMask)
	})

	require.NoError(t, image.UpdateFeatures(FeatureObjectMap|FeatureFastDiff, true))

	t.Run("featuresEnabled", func(t *testing.T) {
		features, err := image.GetFeatures()
		require.NoError(t, err)
		require.Equal(t, FeatureObjectMap|FeatureFastDiff,
			features&(FeatureObjectMap|FeatureFastDiff))

		flags, err := image.GetFlags()
		require.NoError(t, err)
		assert.Equal(t, invalidMask, flags&invalidMask)
	})

	t.Run("readOnlyImage", func(t *testing.T) {
		readOnlyImage, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, readOnlyImage.Close()) })

		flags, err := readOnlyImage.GetFlags()
		require.NoError(t, err)
		assert.Equal(t, invalidMask, flags&invalidMask)
	})

	t.Run("closedImage", func(t *testing.T) {
		closedImage, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		t.Cleanup(func() {
			if closedImage.image != nil {
				require.NoError(t, closedImage.Close())
			}
		})
		require.NoError(t, closedImage.Close())

		flags, err := closedImage.GetFlags()
		assert.ErrorIs(t, err, ErrImageNotOpen)
		assert.Zero(t, flags)
	})
}
