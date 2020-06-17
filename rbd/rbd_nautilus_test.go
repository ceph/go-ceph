// +build !luminous,!mimic

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImagePropertiesNautilus(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer func() { assert.NoError(t, conn.DeletePool(poolname)) }()

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	require.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)
	defer func() { assert.NoError(t, img.Remove()) }()
	defer func() { assert.NoError(t, img.Close()) }()

	_, err = img.GetCreateTimestamp()
	assert.NoError(t, err)

	_, err = img.GetAccessTimestamp()
	assert.NoError(t, err)

	_, err = img.GetModifyTimestamp()
	assert.NoError(t, err)
}

func TestClosedImageNautilus(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, conn.DeletePool(poolname)) }()

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	image, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, image.Remove()) }()

	// close the image
	err = image.Close()
	assert.NoError(t, err)

	// functions should now fail with an RBDError

	_, err = image.GetCreateTimestamp()
	assert.Error(t, err)

	_, err = image.GetAccessTimestamp()
	assert.Error(t, err)

	_, err = image.GetModifyTimestamp()
	assert.Error(t, err)
}
