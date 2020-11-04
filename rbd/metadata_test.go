package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageMetadata(t *testing.T) {
	metadataKey := "mykey"
	metadataValue := "myvalue"

	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	image := GetImage(ioctx, name)
	// Set a metadata key/value on unopen image
	err = image.SetMetadata(metadataKey, metadataValue)
	assert.Equal(t, err, ErrImageNotOpen)
	// Get the metadata value on unopen image
	value, err := image.GetMetadata(metadataKey)
	assert.Equal(t, err, ErrImageNotOpen)
	assert.Equal(t, "", value)
	// Remove the metadata key on unopen image
	err = image.RemoveMetadata(metadataKey)
	assert.Equal(t, err, ErrImageNotOpen)
	// check key is removed on unopen image
	value, err = image.GetMetadata(metadataKey)
	assert.Equal(t, "", value)
	assert.Equal(t, err, ErrImageNotOpen)
	// check ListMetadata for unopened image
	_, err = image.ListMetadata()
	assert.Equal(t, err, ErrImageNotOpen)

	image, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	// Set a metadata key/value
	err = image.SetMetadata(metadataKey, metadataValue)
	assert.NoError(t, err)
	// Get the metadata value
	value, err = image.GetMetadata(metadataKey)
	assert.NoError(t, err)
	assert.Equal(t, metadataValue, value)
	// List metadata
	mm, err := image.ListMetadata()
	assert.NoError(t, err)
	assert.Len(t, mm, 1)
	if assert.Contains(t, mm, metadataKey) {
		assert.Equal(t, mm[metadataKey], metadataValue)
	}
	// Remove the metadata key
	err = image.RemoveMetadata(metadataKey)
	assert.NoError(t, err)
	// check key is removed
	value, err = image.GetMetadata(metadataKey)
	assert.Equal(t, "", value)
	assert.Error(t, err)

	// Set some additional metadata values
	m1 := map[string]string{
		metadataKey:  metadataValue,
		"k2":         "himalayas",
		"matterhorn": "alps",
		"aconcagua":  "andes",
	}
	for k, v := range m1 {
		err = image.SetMetadata(k, v)
		assert.NoError(t, err)
	}
	// Get the metadata value
	value, err = image.GetMetadata(metadataKey)
	assert.NoError(t, err)
	assert.Equal(t, metadataValue, value)
	// List metadata
	mm, err = image.ListMetadata()
	assert.NoError(t, err)
	assert.Len(t, mm, 4)
	assert.Equal(t, m1, mm)

	err = image.Close()
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}
