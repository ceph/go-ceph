//go:build !(octopus || nautilus)
// +build !octopus,!nautilus

package rbd

import (
	"testing"

	"github.com/ceph/go-ceph/rados"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createAndWriteDataToImage(t *testing.T, ioctx *rados.IOContext) string {
	opts := NewRbdImageOptions()
	defer opts.Destroy()

	assert.NoError(t,
		opts.SetUint64(ImageOptionOrder, uint64(testImageOrder)))

	name := GetUUID()
	err := CreateImage(ioctx, name, 1<<25, opts)
	require.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)

	img.CreateSnapshot("snap1")

	_, err = img.WriteAt([]byte("sometimes you feel like a nut"), 0)
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	return name
}

func TestMigration(t *testing.T) {
	conn := radosConnect(t)

	pool := GetUUID()

	err := conn.MakePool(pool)
	require.NoError(t, err)

	ioctx, err := conn.OpenIOContext(pool)
	require.NoError(t, err)

	name := createAndWriteDataToImage(t, ioctx)
	destImage := GetUUID()

	err = MigrationPrepare(ioctx, name, ioctx, destImage, NewRbdImageOptions())
	require.NoError(t, err)

	status, err := MigrationStatus(ioctx, destImage)
	require.NoError(t, err)
	assert.Equal(t, status.State, MigrationImagePrepared)

	err = MigrationExecute(ioctx, destImage)
	require.NoError(t, err)

	status, err = MigrationStatus(ioctx, destImage)
	require.NoError(t, err)
	assert.Equal(t, status.State, MigrationImageExecuted)

	err = MigrationCommit(ioctx, destImage)
	require.NoError(t, err)

	_, err = OpenImage(ioctx, destImage, NoSnapshot)
	assert.NoError(t, err)

	// original image is moved to trash as part of migration prepare
	_, err = OpenImage(ioctx, name, NoSnapshot)
	assert.Error(t, err)

	ioctx.Destroy()
	conn.DeletePool(pool)
	conn.Shutdown()

}

func TestMigrationPrepareImport(t *testing.T) {
	conn := radosConnect(t)

	pool := GetUUID()
	err := conn.MakePool(pool)
	require.NoError(t, err)

	ioctx, err := conn.OpenIOContext(pool)
	require.NoError(t, err)

	name := createAndWriteDataToImage(t, ioctx)
	destImage := GetUUID()
	sourceSpec := `{"type": "native", "snap_name": "snap1", "pool_name":"` + pool + `", "image_name": "` + name + `"}}`

	err = MigrationPrepareImport(sourceSpec, ioctx, destImage, NewRbdImageOptions())
	require.NoError(t, err)

	status, err := MigrationStatus(ioctx, destImage)
	require.NoError(t, err)
	assert.Equal(t, status.State, MigrationImagePrepared)

	img, err := OpenImage(ioctx, destImage, NoSnapshot)
	assert.NoError(t, err)

	err = img.Remove()
	assert.Error(t, err)

	ioctx.Destroy()
	conn.DeletePool(pool)
	conn.Shutdown()

}

func TestMigrationAbort(t *testing.T) {
	conn := radosConnect(t)

	pool := GetUUID()

	err := conn.MakePool(pool)
	require.NoError(t, err)

	ioctx, err := conn.OpenIOContext(pool)
	require.NoError(t, err)

	name := createAndWriteDataToImage(t, ioctx)
	destImage := GetUUID()

	err = MigrationPrepare(ioctx, name, ioctx, destImage, NewRbdImageOptions())
	require.NoError(t, err)

	status, err := MigrationStatus(ioctx, destImage)
	require.NoError(t, err)
	assert.Equal(t, status.State, MigrationImagePrepared)

	err = MigrationAbort(ioctx, destImage)
	require.NoError(t, err)

	_, err = OpenImage(ioctx, destImage, NoSnapshot)
	assert.Error(t, err)

	// original image is retrievable
	_, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(pool)
	conn.Shutdown()

}
