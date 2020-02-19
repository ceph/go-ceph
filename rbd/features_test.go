package rbd

import (
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
	err = options.SetUint64(RbdImageOptionFeatures, FeatureLayering|FeatureStripingV2)
	require.NoError(t, err)
	// FeatureStripingV2 only works with additional arguments
	err = options.SetUint64(RbdImageOptionStripeUnit, 1024*1024)
	require.NoError(t, err)
	err = options.SetUint64(RbdImageOptionStripeCount, 4)
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
}
