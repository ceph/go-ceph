package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSnapshot(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	_, err = img.CreateSnapshot("mysnap")
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	snapImage, err := OpenImage(ioctx, name, "mysnap")
	assert.NoError(t, err)

	err = snapImage.Close()
	assert.NoError(t, err)

	img2, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	snapshot := img2.GetSnapshot("mysnap")

	err = snapshot.Remove()
	assert.NoError(t, err)

	err = img2.Close()
	assert.NoError(t, err)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestErrorSnapshotNoName(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	// this actually works for some reason?!
	snapshot, err := img.CreateSnapshot("")
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	err = snapshot.Remove()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Rollback()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Protect()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Unprotect()
	assert.Equal(t, err, ErrSnapshotNoName)

	_, err = snapshot.IsProtected()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Set()
	assert.Equal(t, err, ErrSnapshotNoName)

	// image can not be removed as the snapshot still exists
	// err = img.Remove()
	// assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}
