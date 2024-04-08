package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageResize2(t *testing.T) {
	cc := 0
	cb := func(_, total uint64, v interface{}) int {
		cc++
		val := v.(int)
		assert.Equal(t, 0, val)
		assert.Equal(t, uint64(2), total)
		return 0
	}

	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	reqSize := uint64(1024 * 1024 * 4) // 4MB
	err = quickCreate(ioctx, name, reqSize, testImageOrder)
	assert.NoError(t, err)

	image, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	size, err := image.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, size, reqSize)

	newReqSize := reqSize * 2

	// Test normal resize (no shrinking allowed)
	err = image.Resize2(newReqSize, false, cb, nil)
	assert.NoError(t, err)

	size, err = image.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, size, newReqSize)

	// Resize to a smaller size with shrinking allowed
	err = image.Resize2(reqSize, true, cb, 0)
	assert.NoError(t, err)

	// Attempt to resize to a smaller size with shrinking disallowed (should error)
	err = image.Resize2(reqSize-1024*1024, false, cb, 0)
	assert.Error(t, err)

	err = image.Close()
	assert.NoError(t, err)

	err = image.Resize2(newReqSize, false, cb, 0)
	assert.Error(t, err) // Expect an error since the image is not open/

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}
