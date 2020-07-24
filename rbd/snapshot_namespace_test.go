// +build !luminous

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSnapNamespaceType(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	imageName := "parent"
	snapName := "mySnap"
	options := NewRbdImageOptions()
	defer options.Destroy()
	err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
	assert.NoError(t, err)
	err = options.SetUint64(ImageOptionFeatures, 1)
	assert.NoError(t, err)

	err = CreateImage(ioctx, imageName, testImageSize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, imageName, NoSnapshot)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, img.Close())
		assert.NoError(t, img.Remove())
	}()

	snapshot, err := img.CreateSnapshot(snapName)
	assert.NoError(t, err)

	snapInfoList, err := img.GetSnapshotNames()
	assert.NoError(t, err)

	snapInfo := snapInfoList[0]
	assert.Equal(t, snapInfo.Name, snapName)

	t.Run("SnapNamespaceTypeInvalidArgs", func(t *testing.T) {
		validImageName := GetUUID()
		validImg, err := Create(ioctx, validImageName, testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, validImg.Remove())
		}()

		validImg = GetImage(ioctx, validImageName)
		// Closed image and a snapshot ID which doesn't belong to this image.
		_, err = validImg.GetSnapNamespaceType(snapInfo.Id)
		assert.Error(t, err)

		// Open image but invalid snap ID.
		validImg, err = OpenImage(ioctx, validImageName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, validImg.Close())
		}()

		_, err = validImg.GetSnapNamespaceType(uint64(22))
		assert.Error(t, err)

		// With non-existing image.
		invalidImageName := GetUUID()
		invalidImg := GetImage(ioctx, invalidImageName)
		_, err = invalidImg.GetSnapNamespaceType(snapInfo.Id)
		assert.Error(t, err)
	})

	t.Run("SnapNamespaceTypeUser", func(t *testing.T) {
		nsType, err := img.GetSnapNamespaceType(snapInfo.Id)
		assert.NoError(t, err)
		assert.Equal(t, nsType, SnapNamespaceTypeUser)
	})

	t.Run("SnapNamespaceTypeTrash", func(t *testing.T) {
		cloneName := "myClone"
		optionsClone := NewRbdImageOptions()
		defer optionsClone.Destroy()
		err := optionsClone.SetUint64(ImageOptionCloneFormat, 2)
		assert.NoError(t, err)

		// Create a clone of the image using the same snapshot.
		err = CloneImage(ioctx, imageName, snapName, ioctx, cloneName, optionsClone)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, RemoveImage(ioctx, cloneName)) }()

		// Once clone is created, remove the snapshot.
		err = snapshot.Remove()
		assert.NoError(t, err)

		// Snapshot would move to the trash because linked clone is still there.
		nsType, err := img.GetSnapNamespaceType(snapInfo.Id)
		assert.NoError(t, err)
		assert.Equal(t, nsType, SnapNamespaceTypeTrash)
	})
}

func TestGetSnapTrashNamespace(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	imageName := "parent"
	snapName := "mySnap"
	cloneName := "myClone"
	options := NewRbdImageOptions()
	defer options.Destroy()
	err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
	assert.NoError(t, err)
	err = options.SetUint64(ImageOptionFeatures, 1)
	assert.NoError(t, err)

	err = CreateImage(ioctx, imageName, testImageSize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, imageName, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Close())
		assert.NoError(t, img.Remove())
	}()

	snapshot, err := img.CreateSnapshot(snapName)
	assert.NoError(t, err)

	optionsClone := NewRbdImageOptions()
	defer optionsClone.Destroy()
	err = optionsClone.SetUint64(ImageOptionCloneFormat, 2)
	assert.NoError(t, err)

	// Create a clone of the image using the same snapshot.
	err = CloneImage(ioctx, imageName, snapName, ioctx, cloneName, optionsClone)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, RemoveImage(ioctx, cloneName)) }()

	// Check the name of the snapshot.
	snapInfoList, err := img.GetSnapshotNames()
	assert.NoError(t, err)
	snapInfo := snapInfoList[0]
	assert.Equal(t, snapInfo.Name, snapName)

	// Remove the snapshot.
	err = snapshot.Remove()
	assert.NoError(t, err)

	// Snapshot would move to the trash because linked clone is still there.
	nsType, err := img.GetSnapNamespaceType(snapInfo.Id)
	assert.NoError(t, err)
	assert.Equal(t, nsType, SnapNamespaceTypeTrash)

	// Get the snap info again, name would have changed.
	newSnapInfoList, err := img.GetSnapshotNames()
	assert.NoError(t, err)
	newSnapInfo := newSnapInfoList[0]
	assert.NotEqual(t, newSnapInfo.Name, snapName)
	// ID would have remained same.
	assert.Equal(t, snapInfo.Id, newSnapInfo.Id)

	// Get the original name.
	origSnapName, err := img.GetSnapTrashNamespace(newSnapInfo.Id)
	assert.NoError(t, err)
	assert.Equal(t, snapName, origSnapName)

	invalidSnapID := uint64(22)

	t.Run("InvalidSnapID", func(t *testing.T) {
		_, err := img.GetSnapTrashNamespace(invalidSnapID)
		assert.Error(t, err)
	})

	t.Run("InvalidImage", func(t *testing.T) {
		invalidImgName := GetUUID()
		invalidImg := GetImage(ioctx, invalidImgName)
		_, err := invalidImg.GetSnapTrashNamespace(invalidSnapID)
		assert.Error(t, err)
	})
}
