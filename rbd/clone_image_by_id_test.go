//go:build !(nautilus || octopus || pacific || quincy || reef) && ceph_preview

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneImageByID(t *testing.T) {
	// tests are done as subtests to avoid creating pools, images, etc
	// over and over again.
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

	// create a group, some images, and add images to the group
	gname := "snapme"
	err = GroupCreate(ioctx, gname)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, GroupRemove(ioctx, gname))
	}()

	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	defer options.Destroy()

	name1 := GetUUID()
	err = CreateImage(ioctx, name1, testImageSize, options)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, RemoveImage(ioctx, name1))
	}()

	name2 := GetUUID()
	err = CreateImage(ioctx, name2, testImageSize, options)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, RemoveImage(ioctx, name2))
	}()

	err = GroupImageAdd(ioctx, gname, ioctx, name1)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, GroupImageRemove(ioctx, gname, ioctx, name1))
	}()

	err = GroupImageAdd(ioctx, gname, ioctx, name2)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, GroupImageRemove(ioctx, gname, ioctx, name2))
	}()

	t.Run("CloneFromSnapshot", func(t *testing.T) {
		cloneName := "child"
		optionsClone := NewRbdImageOptions()
		defer optionsClone.Destroy()
		err := optionsClone.SetUint64(ImageOptionCloneFormat, 2)
		assert.NoError(t, err)

		// Get the snapID
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()

		snapName := "mysnap"
		snapshot, err := img.CreateSnapshot(snapName)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, snapshot.Remove())
		}()

		snapInfos, err := img.GetSnapshotNames()
		assert.NoError(t, err)
		require.Equal(t, 1, len(snapInfos))

		snapID := snapInfos[0].Id

		// Create a clone of the image using the snapshot.
		err = CloneImageByID(ioctx, name1, snapID, ioctx, cloneName, optionsClone)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, RemoveImage(ioctx, cloneName)) }()

		imgNew, err := OpenImage(ioctx, cloneName, NoSnapshot)
		defer func() {
			assert.NoError(t, imgNew.Close())
		}()
		assert.NoError(t, err)

		parentInfo, err := imgNew.GetParent()
		assert.NoError(t, err)
		assert.Equal(t, parentInfo.Image.ImageName, name1)
		assert.Equal(t, parentInfo.Image.PoolName, poolname)
		assert.False(t, parentInfo.Image.Trash)
		assert.Equal(t, parentInfo.Snap.SnapName, snapName)
		assert.Equal(t, parentInfo.Snap.ID, snapID)
	})

	t.Run("CloneFromGroupSnap", func(t *testing.T) {
		err := GroupSnapCreate(ioctx, gname, "groupsnap")
		assert.NoError(t, err)

		cloneName := "img-clone"
		optionsClone := NewRbdImageOptions()
		defer optionsClone.Destroy()
		err = optionsClone.SetUint64(ImageOptionCloneFormat, 2)
		assert.NoError(t, err)

		// Get the snapID of the image
		img, err := OpenImageReadOnly(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		defer img.Close()

		snapInfos, err := img.GetSnapshotNames()
		assert.NoError(t, err)
		require.Equal(t, 1, len(snapInfos))

		snapID := snapInfos[0].Id

		// Create a clone of the image using the snapshot.
		err = CloneImageByID(ioctx, name1, snapID, ioctx, cloneName, optionsClone)
		assert.NoError(t, err)
		defer func() { assert.NoError(t, RemoveImage(ioctx, cloneName)) }()

		imgNew, err := OpenImage(ioctx, cloneName, NoSnapshot)
		defer func() {
			assert.NoError(t, imgNew.Close())
		}()
		assert.NoError(t, err)

		parentInfo, err := imgNew.GetParent()
		assert.NoError(t, err)
		assert.Equal(t, parentInfo.Image.ImageName, name1)
		assert.Equal(t, parentInfo.Snap.ID, snapID)
		assert.Equal(t, parentInfo.Image.PoolName, poolname)
		assert.False(t, parentInfo.Image.Trash)

		err = GroupSnapRemove(ioctx, gname, "groupsnap")
		assert.NoError(t, err)
	})
}
