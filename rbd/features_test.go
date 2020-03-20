package rbd

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFeatures(t *testing.T) {
	conn := radosConnect(t)
	require.NotNil(t, conn)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()

	options := NewRbdImageOptions()
	err = options.SetUint64(ImageOptionFeatures, FeatureLayering|FeatureStripingV2)
	require.NoError(t, err)
	// FeatureStripingV2 only works with additional arguments
	err = options.SetUint64(ImageOptionStripeUnit, 1024*1024)
	require.NoError(t, err)
	err = options.SetUint64(ImageOptionStripeCount, 4)
	require.NoError(t, err)

	err = CreateImage(ioctx, name, 16*1024*1024, options)
	require.NoError(t, err)
	defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

	image, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, image.Close()) }()

	features, err := image.GetFeatures()
	assert.NoError(t, err)

	t.Run("compareBits", func(t *testing.T) {
		hasLayering := (features & FeatureLayering) == FeatureLayering
		hasStripingV2 := (features & FeatureStripingV2) == FeatureStripingV2
		assert.True(t, hasLayering, "FeatureLayering is not set")
		assert.True(t, hasStripingV2, "FeatureStripingV2 is not set")
	})

	t.Run("compareFeatureSet", func(t *testing.T) {
		fs := FeatureSet(features)
		assert.Contains(t, fs.Names(), FeatureNameLayering)
		assert.Contains(t, fs.Names(), FeatureNameStripingV2)
	})
}

func TestFeatureSet(t *testing.T) {
	fsBits := FeatureSet(FeatureExclusiveLock | FeatureDeepFlatten)
	fsNames := FeatureSetFromNames([]string{FeatureNameExclusiveLock, FeatureNameDeepFlatten})
	assert.Equal(t, fsBits, fsNames)

	fsBitsSorted := fsBits.Names()
	sort.Strings(fsBitsSorted)

	fsNamesSorted := fsNames.Names()
	sort.Strings(fsNamesSorted)

	assert.Equal(t, fsBitsSorted, fsNamesSorted)
}

func TestUpdateFeatures(t *testing.T) {
	conn := radosConnect(t)
	require.NotNil(t, conn)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()

	options := NewRbdImageOptions()
	// test with FeatureExclusiveLock as that is mutable
	err = options.SetUint64(ImageOptionFeatures, FeatureExclusiveLock)
	require.NoError(t, err)

	err = CreateImage(ioctx, name, 16*1024*1024, options)
	require.NoError(t, err)
	defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

	image, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)
	defer func() { assert.NoError(t, image.Close()) }()

	t.Run("imageNotOpen", func(t *testing.T) {
		img, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, img)

		err = img.Close()
		require.NoError(t, err)

		err = img.UpdateFeatures(FeatureExclusiveLock, false)
		assert.Equal(t, err, ErrImageNotOpen)
	})

	t.Run("verifyFeatureEnabled", func(t *testing.T) {
		features, err := image.GetFeatures()
		require.NoError(t, err)

		hasExclusiveLock := (features & FeatureExclusiveLock) == FeatureExclusiveLock
		require.True(t, hasExclusiveLock, "FeatureExclusiveLock is not set")
	})

	t.Run("disableFeature", func(t *testing.T) {
		err = image.UpdateFeatures(FeatureExclusiveLock, false)
		require.NoError(t, err)

		features, err := image.GetFeatures()
		require.NoError(t, err)

		hasExclusiveLock := (features & FeatureExclusiveLock) == FeatureExclusiveLock
		require.False(t, hasExclusiveLock, "FeatureExclusiveLock is set")
	})

	t.Run("enableFeature", func(t *testing.T) {
		err = image.UpdateFeatures(FeatureExclusiveLock, true)
		require.NoError(t, err)

		features, err := image.GetFeatures()
		require.NoError(t, err)

		hasExclusiveLock := (features & FeatureExclusiveLock) == FeatureExclusiveLock
		require.True(t, hasExclusiveLock, "FeatureExclusiveLock is not set")
	})
}
