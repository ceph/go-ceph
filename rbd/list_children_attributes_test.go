package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListChildrenAttributes(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer conn.DeletePool(poolName)

	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)
	defer ioctx.Destroy()

	namespace := "ns01"
	err = NamespaceCreate(ioctx, namespace)
	assert.NoError(t, err)

	ioctx.SetNamespace(namespace)

	parentName := "parent"
	img, err := Create(ioctx, parentName, testImageSize, testImageOrder, 1)
	assert.NoError(t, err)
	defer img.Remove()

	img, err = OpenImage(ioctx, parentName, NoSnapshot)
	assert.NoError(t, err)
	defer img.Close()

	snapName := "snapshot"
	snapshot, err := img.CreateSnapshot(snapName)
	assert.NoError(t, err)
	defer snapshot.Remove()

	err = snapshot.Protect()
	assert.NoError(t, err)

	snapImg, err := OpenImage(ioctx, parentName, snapName)
	assert.NoError(t, err)
	defer snapImg.Close()

	// ensure no children prior to clone
	result, err := snapImg.ListChildrenAttributes()
	assert.NoError(t, err)
	assert.Equal(t, len(result), 0, "List should be empty before cloning")

	//create first child image
	childImage01 := "childImage01"
	_, err = img.Clone(snapName, ioctx, childImage01, 1, testImageOrder)
	assert.NoError(t, err)

	//retrieve and validate child image properties
	result, err = snapImg.ListChildrenAttributes()
	assert.NoError(t, err)
	assert.Equal(t, len(result), 1, "List should contain one child image")

	assert.Equal(t, childImage01, result[0].ImageName)
	assert.NotNil(t, result[0].ImageID)
	assert.NotNil(t, result[0].PoolID)
	assert.Equal(t, poolName, result[0].PoolName)
	assert.Equal(t, namespace, result[0].PoolNamespace)
	assert.False(t, result[0].Trash, "Newly cloned image should not be in trash")

	//parent image cannot be deleted while having children attached to it
	err = img.Remove()
	assert.Error(t, err, "Expected an error but got nil")

	childImg1, err := OpenImage(ioctx, childImage01, NoSnapshot)
	assert.NoError(t, err)
	defer childImg1.Close()

	//trash the image and validate
	err = childImg1.Trash(0)
	assert.NoError(t, err, "Failed to move child image to trash")

	result, err = snapImg.ListChildrenAttributes()
	assert.NoError(t, err)
	assert.True(t, result[0].Trash, "Child image should be marked as trashed")

	//validate for multiple clones by creating second child image
	childImage02 := "childImage02"
	_, err = img.Clone(snapName, ioctx, childImage02, 1, testImageOrder)
	assert.NoError(t, err)

	result, err = snapImg.ListChildrenAttributes()
	assert.NoError(t, err)
	require.Len(t, result, 2, "List should contain two child images")
	assert.Equal(t, childImage02, result[1].ImageName)

	//the first image is trashed where as the second is not
	expectedChildren := map[string]bool{
		childImage01: true,
		childImage02: false,
	}

	for _, child := range result {
		exists := expectedChildren[child.ImageName]
		assert.Equal(t, exists, child.Trash)
	}

	childImg2, err := OpenImage(ioctx, childImage02, NoSnapshot)
	require.NoError(t, err, "Failed to open cloned child image")
	defer childImg2.Close()

	//flattening the image should detach it from the parent and remove it from ListChildrenAttributes
	err = childImg2.Flatten()
	assert.NoError(t, err, "Failed to flatten cloned child image")

	result, err = snapImg.ListChildrenAttributes()
	assert.NoError(t, err)
	assert.Equal(t, len(result), 1, "List should not contain the second image after flattening the clone")
	assert.NotEqual(t, childImage02, result[0].ImageName)
}

func TestCloneInDifferentPool(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	//create two pools:poolA(parent) and poolB(child)
	poolA := GetUUID()
	err := conn.MakePool(poolA)
	require.NoError(t, err)
	defer conn.DeletePool(poolA)

	poolB := GetUUID()
	err = conn.MakePool(poolB)
	require.NoError(t, err)
	defer conn.DeletePool(poolB)

	//ensure that both the pools are not the same
	assert.NotEqual(t, poolA, poolB)

	ioctxA, err := conn.OpenIOContext(poolA)
	require.NoError(t, err)
	defer ioctxA.Destroy()
	ioctxB, err := conn.OpenIOContext(poolB)
	require.NoError(t, err)
	defer ioctxB.Destroy()

	//create a parent image in poolA
	parentName := "parent-image"
	img, err := Create(ioctxA, parentName, testImageSize, testImageOrder, 1)
	assert.NoError(t, err)

	img, err = OpenImage(ioctxA, parentName, NoSnapshot)
	assert.NoError(t, err)
	defer img.Close()

	snapName := "snap01"
	snapshot, err := img.CreateSnapshot(snapName)
	assert.NoError(t, err)
	defer snapshot.Remove()

	err = snapshot.Protect()
	assert.NoError(t, err)

	snapImg, err := OpenImage(ioctxA, parentName, snapName)
	assert.NoError(t, err)
	defer snapImg.Close()

	//create a child image in poolB
	childImageName := "child-image"
	_, err = img.Clone(snapName, ioctxB, childImageName, 1, testImageOrder)
	assert.NoError(t, err, "Failed to clone image into poolB")

	//verify and validate properties of the child image
	result, err := snapImg.ListChildrenAttributes()
	assert.NoError(t, err)
	require.Len(t, result, 1, "List should contain one child image")

	child := result[0]
	assert.Equal(t, childImageName, child.ImageName, "Child image name should match")
	assert.Equal(t, poolB, child.PoolName, "Child image should be in poolB")
}
