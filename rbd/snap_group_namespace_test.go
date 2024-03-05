//go:build ceph_preview

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSnapGroupNamespace(t *testing.T) {
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
	groupName := "myGroup"
	groupSnapName := "myGroupSnap"
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

	err = GroupCreate(ioctx, groupName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, GroupRemove(ioctx, groupName))
	}()

	err = GroupImageAdd(ioctx, groupName, ioctx, imageName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, GroupImageRemove(ioctx, groupName, ioctx, imageName))
	}()

	err = GroupSnapCreate(ioctx, groupName, groupSnapName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, GroupSnapRemove(ioctx, groupName, groupSnapName))
	}()

	// Take the details of the 1st snapshot of the image.
	snapInfoList, err := img.GetSnapshotNames()
	assert.NoError(t, err)
	snapInfo := snapInfoList[0]
	assert.Positive(t, snapInfo.Id)
	assert.Regexp(t, "^\\.group\\.", snapInfo.Name)

	// The snapshot is expected to be in the 'group' namespace.
	nsType, err := img.GetSnapNamespaceType(snapInfo.Id)
	assert.NoError(t, err)
	assert.Equal(t, nsType, SnapNamespaceTypeGroup)

	// Get the info from the snapshot in the group.
	sgn, err := img.GetSnapGroupNamespace(snapInfo.Id)
	assert.NoError(t, err)
	require.NotNil(t, sgn)
	assert.Equal(t, groupName, sgn.GroupName)

	// Negative testing follows.
	invalidSnapID := uint64(22)

	t.Run("InvalidSnapID", func(t *testing.T) {
		_, err := img.GetSnapGroupNamespace(invalidSnapID)
		assert.Error(t, err)
	})

	t.Run("InvalidImage", func(t *testing.T) {
		invalidImgName := GetUUID()
		invalidImg := GetImage(ioctx, invalidImgName)
		_, err := invalidImg.GetSnapGroupNamespace(invalidSnapID)
		assert.Error(t, err)
	})
}
