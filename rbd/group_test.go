package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCreateRemove(t *testing.T) {
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

	err = GroupCreate(ioctx, "group1")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "group1")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "group2")
	assert.NoError(t, err)

	err = GroupCreate(ioctx, "group2")
	assert.NoError(t, err)
	err = GroupCreate(ioctx, "group")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "group2")
	assert.NoError(t, err)
}

func TestGroupRename(t *testing.T) {
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

	err = GroupCreate(ioctx, "group1")
	assert.NoError(t, err)

	err = GroupRename(ioctx, "group1", "club1")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "club1")
	assert.NoError(t, err)

	// unlike remove, rename does return an error if the src name
	// doesn't exist
	err = GroupRename(ioctx, "club1", "nowhere")
	assert.Error(t, err)
}

func TestGroupList(t *testing.T) {
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

	err = GroupCreate(ioctx, "uno")
	assert.NoError(t, err)
	err = GroupCreate(ioctx, "dos")
	assert.NoError(t, err)
	err = GroupCreate(ioctx, "tres")
	assert.NoError(t, err)

	l, err := GroupList(ioctx)
	assert.NoError(t, err)
	if assert.Len(t, l, 3) {
		assert.Contains(t, l, "uno")
		assert.Contains(t, l, "dos")
		assert.Contains(t, l, "tres")
	}

	err = GroupRemove(ioctx, "uno")
	assert.NoError(t, err)
	err = GroupRemove(ioctx, "dos")
	assert.NoError(t, err)
	err = GroupRemove(ioctx, "tres")
	assert.NoError(t, err)

	l, err = GroupList(ioctx)
	assert.NoError(t, err)
	assert.Len(t, l, 0)

	// test that GroupList panics if passed a nil ioctx
	assert.Panics(t, func() {
		GroupList(nil)
	})
}

func TestGroupImageAdd(t *testing.T) {
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
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	require.NoError(t, err)

	err = GroupCreate(ioctx, "grone")
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, "grone", ioctx, name)
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, "badGroup", ioctx, name)
	assert.Error(t, err)

	assert.Panics(t, func() {
		GroupImageAdd(nil, "invalid", ioctx, "foobar")
	})
	assert.Panics(t, func() {
		GroupImageAdd(ioctx, "invalid", nil, "foobar")
	})
}

func TestGroupImageRemove(t *testing.T) {
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
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	require.NoError(t, err)

	err = GroupCreate(ioctx, "grone")
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, "grone", ioctx, name)
	assert.NoError(t, err)

	err = GroupImageRemove(ioctx, "grone", ioctx, name)
	assert.NoError(t, err)

	err = GroupImageRemove(ioctx, "badGroup", ioctx, name)
	assert.Error(t, err)

	assert.Panics(t, func() {
		GroupImageRemove(nil, "invalid", ioctx, "foobar")
	})
	assert.Panics(t, func() {
		GroupImageRemove(ioctx, "invalid", nil, "foobar")
	})
}

func TestGroupImageRemoveByID(t *testing.T) {
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
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	require.NoError(t, err)

	err = GroupCreate(ioctx, "grone")
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, "grone", ioctx, name)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	imageID, err := img.GetId()
	assert.NoError(t, err)
	err = img.Close()
	assert.NoError(t, err)

	err = GroupImageRemoveByID(ioctx, "grone", ioctx, imageID)
	assert.NoError(t, err)

	err = GroupImageRemoveByID(ioctx, "badGroup", ioctx, imageID)
	assert.Error(t, err)

	assert.Panics(t, func() {
		GroupImageRemoveByID(nil, "invalid", ioctx, "foobar")
	})
	assert.Panics(t, func() {
		GroupImageRemoveByID(ioctx, "invalid", nil, "foobar")
	})
}

func TestGroupImageList(t *testing.T) {
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

	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))

	name1 := GetUUID()
	err = CreateImage(ioctx, name1, testImageSize, options)
	require.NoError(t, err)

	name2 := GetUUID()
	err = CreateImage(ioctx, name2, testImageSize, options)
	require.NoError(t, err)

	err = GroupCreate(ioctx, "grone")
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, "grone", ioctx, name1)
	assert.NoError(t, err)

	gimgs, err := GroupImageList(ioctx, "grone")
	assert.NoError(t, err)
	assert.NotNil(t, gimgs)
	if assert.Len(t, gimgs, 1) {
		assert.Equal(t, name1, gimgs[0].Name)
		assert.Equal(t, GroupImageStateAttached, gimgs[0].State)
	}

	err = GroupImageAdd(ioctx, "grone", ioctx, name2)
	assert.NoError(t, err)

	gimgs, err = GroupImageList(ioctx, "grone")
	assert.NoError(t, err)
	assert.NotNil(t, gimgs)
	if assert.Len(t, gimgs, 2) {
		// order is not assured!
		names := []string{}
		for _, gimg := range gimgs {
			names = append(names, gimg.Name)
		}
		assert.Contains(t, names, name1)
		assert.Contains(t, names, name2)
	}

	err = GroupImageRemove(ioctx, "grone", ioctx, name1)
	assert.NoError(t, err)

	gimgs, err = GroupImageList(ioctx, "grone")
	assert.NoError(t, err)
	assert.NotNil(t, gimgs)
	if assert.Len(t, gimgs, 1) {
		assert.Equal(t, name2, gimgs[0].Name)
		assert.Equal(t, GroupImageStateAttached, gimgs[0].State)
	}

	assert.Panics(t, func() {
		GroupImageList(nil, "foo")
	})
}

func TestGroupImageGetGroup(t *testing.T) {
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

	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))

	name1 := GetUUID()
	err = CreateImage(ioctx, name1, testImageSize, options)
	require.NoError(t, err)

	name2 := GetUUID()
	err = CreateImage(ioctx, name2, testImageSize, options)
	require.NoError(t, err)

	err = GroupCreate(ioctx, "grone")
	assert.NoError(t, err)

	err = GroupImageAdd(ioctx, "grone", ioctx, name1)
	assert.NoError(t, err)

	img1, err := OpenImage(ioctx, name1, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img1.Close())
	}()

	gi1, err := img1.GetGroup()
	assert.NoError(t, err)
	assert.Equal(t, "grone", gi1.Name)

	img2, err := OpenImage(ioctx, name2, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img2.Close())
	}()

	gi2, err := img2.GetGroup()
	assert.NoError(t, err)
	assert.Equal(t, "", gi2.Name)

	t.Run("invalidImage", func(t *testing.T) {
		x := &Image{}
		_, err := x.GetGroup()
		assert.Error(t, err)
		assert.Panics(t, func() {
			var x *Image
			x.GetGroup()
		})
	})
}

func TestGetGroupMirrorStatus(t *testing.T) {
	conn := radosConnect(t)
	require.NotNil(t, conn)
	defer conn.Shutdown()

	// poolname := GetUUID()
	poolname := "data"
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	// defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	status, err := GetGlobalMirrorGroupStatus(ioctx, "grp1")
	assert.NoError(t, err)
	assert.NotNil(t, status)
}
