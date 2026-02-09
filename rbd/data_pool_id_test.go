package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataPoolID(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	dataPoolname := GetUUID()
	err = conn.MakePool(dataPoolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t, options.SetString(ImageOptionDataPool, dataPoolname))

	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	workingImage, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	dataPoolId, err := conn.GetPoolByName(dataPoolname)
	assert.NoError(t, err)

	t.Run("GetDataPoolId", func(t *testing.T) {
		id, err := workingImage.GetDataPoolID()
		assert.NoError(t, err)
		assert.Equal(t, dataPoolId, id)
	})

	singlePoolImageName := GetUUID()
	singlePoolImageOptions := NewRbdImageOptions()
	err = CreateImage(ioctx, singlePoolImageName, testImageSize, singlePoolImageOptions)
	assert.NoError(t, err)

	singlePoolImage, err := OpenImage(ioctx, singlePoolImageName, NoSnapshot)
	assert.NoError(t, err)

	t.Run("GetDataPoolIdNoDataPool", func(t *testing.T) {
		expectedPoolId, err := conn.GetPoolByName(poolname)
		assert.NoError(t, err)

		id, err := singlePoolImage.GetDataPoolID()

		assert.NoError(t, err)
		assert.Equal(t, expectedPoolId, id)
	})

	assert.NoError(t, workingImage.Close())
	err = workingImage.Remove()
	assert.NoError(t, err)

	assert.NoError(t, singlePoolImage.Close())
	err = singlePoolImage.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.DeletePool(dataPoolname)
	conn.Shutdown()
}
