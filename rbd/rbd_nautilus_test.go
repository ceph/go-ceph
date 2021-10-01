//go:build !luminous && !mimic
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

	// functions should now fail with an rbdError

	_, err = image.GetCreateTimestamp()
	assert.Error(t, err)

	_, err = image.GetAccessTimestamp()
	assert.Error(t, err)

	_, err = image.GetModifyTimestamp()
	assert.Error(t, err)
}

func TestSparsify(t *testing.T) {
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

		err = img.Sparsify(4096)
		assert.NoError(t, err)
	})

	t.Run("invalidValue", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, img.Close()) }()

		err = img.Sparsify(1024)
		assert.Error(t, err)
	})

	t.Run("closedImage", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())

		err = img.Sparsify(1024)
		assert.Error(t, err)
	})
}

func TestGetParent(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	assert.NoError(t, err)
	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)

	defer func() {
		ioctx.Destroy()
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	imgName := "parent"
	img, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Remove())
	}()

	img, err = OpenImage(ioctx, imgName, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Close())
	}()

	snapName := "mysnap"
	snapshot, err := img.CreateSnapshot(snapName)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, snapshot.Remove())
	}()

	t.Run("ParentNotAvailable", func(t *testing.T) {
		_, err := img.GetParent()
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("ParentAvailable", func(t *testing.T) {
		cloneName := "child"
		optionsClone := NewRbdImageOptions()
		defer optionsClone.Destroy()
		err := optionsClone.SetUint64(ImageOptionCloneFormat, 2)
		assert.NoError(t, err)

		// Create a clone of the image using the snapshot.
		err = CloneImage(ioctx, imgName, snapName, ioctx, cloneName, optionsClone)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, RemoveImage(ioctx, cloneName)) }()

		imgNew, err := OpenImage(ioctx, cloneName, NoSnapshot)
		defer func() {
			assert.NoError(t, imgNew.Close())
		}()
		assert.NoError(t, err)

		parentInfo, err := imgNew.GetParent()
		assert.NoError(t, err)
		assert.Equal(t, parentInfo.Image.ImageName, imgName)
		assert.Equal(t, parentInfo.Snap.SnapName, snapName)
		assert.Equal(t, parentInfo.Image.PoolName, poolName)
		// TODO: add a comaprison for snap ID
	})

	t.Run("ClosedImage", func(t *testing.T) {
		closedImg, err := Create(ioctx, "someImage", testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, closedImg.Remove())
		}()
		_, err = closedImg.GetParent()
		assert.Error(t, err)
	})
}

func TestSetSnapByID(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	assert.NoError(t, err)
	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)

	defer func() {
		ioctx.Destroy()
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	t.Run("happyPath", func(t *testing.T) {
		imgName := "Hogwarts"
		img, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Remove())
		}()

		img, err = OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()
		bytesIn := []byte("Gryffindor")
		in, err := img.Write(bytesIn)
		assert.NoError(t, err)
		assert.Equal(t, in, len(bytesIn))

		snapName := "myPensieve"
		snapshot, err := img.CreateSnapshot(snapName)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, snapshot.Remove())
		}()

		// overwrite
		_, err = img.Seek(0, SeekSet)
		assert.NoError(t, err)
		bytesOver := []byte("Slytherin")
		over, err := img.Write(bytesOver)
		assert.NoError(t, err)
		assert.Equal(t, over, len(bytesOver))

		snapInfo, err := img.GetSnapshotNames()
		assert.NoError(t, err)
		snapID := snapInfo[0].Id
		assert.Equal(t, snapName, snapInfo[0].Name)

		err = img.SetSnapByID(snapID)
		assert.NoError(t, err)

		// read
		_, err = img.Seek(0, SeekSet)
		assert.NoError(t, err)
		bytesOut := make([]byte, len(bytesIn))
		out, err := img.Read(bytesOut)
		assert.NoError(t, err)
		assert.Equal(t, out, len(bytesOut))
		assert.Equal(t, bytesIn, bytesOut)
	})

	t.Run("ClosedImage", func(t *testing.T) {
		closedImg, err := Create(ioctx, "someImage", testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, closedImg.Remove())
		}()

		var snapID uint64
		snapID = 22
		err = closedImg.SetSnapByID(snapID)
		assert.Error(t, err)
		assert.Equal(t, err, ErrImageNotOpen)
	})

	t.Run("invalidSnapID", func(t *testing.T) {
		imgName := "someImage"
		img, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Remove())
		}()

		img, err = OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		var snapID uint64
		snapID = 22
		err = img.SetSnapByID(snapID)
		assert.Error(t, err)
	})
}
