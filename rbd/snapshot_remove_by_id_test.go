//go:build ceph_preview
// +build ceph_preview

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveSnapByID(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)

	defer func() {
		ioctx.Destroy()
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	t.Run("happyPath", func(t *testing.T) {
		imgName := "myImage"
		img, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Remove())
		}()

		img, err = OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		snapName := "mySnap"
		snapshot, err := img.CreateSnapshot(snapName)
		require.NoError(t, err)
		defer func() {
			assert.Error(t, snapshot.Remove())
		}()

		snapID, err := img.GetSnapID(snapshot.name)
		require.NoError(t, err)
		require.NoError(t, img.RemoveSnapByID(snapID))
	})

	t.Run("closedImage", func(t *testing.T) {
		imgName := "myImage"
		closedImg, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, closedImg.Remove())
		}()

		img, err := OpenImage(ioctx, imgName, NoSnapshot)
		require.NoError(t, err)
		snapName := "mySnap"
		snapshot, err := img.CreateSnapshot(snapName)
		require.NoError(t, err)

		snapID, err := img.GetSnapID(snapshot.name)
		require.NoError(t, err)

		// close the image
		require.NoError(t, img.Close())

		// try to remove the snapshot
		require.Error(t, img.RemoveSnapByID(snapID))

		// Open the image for snapshot removal
		img, err = OpenImage(ioctx, imgName, NoSnapshot)
		require.NoError(t, err)
		require.NoError(t, img.RemoveSnapByID(snapID))
		require.NoError(t, img.Close())
	})
}
