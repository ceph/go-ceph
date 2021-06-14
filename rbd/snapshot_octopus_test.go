// +build !nautilus

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnapIDFunctions(t *testing.T) {
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
		imgName := "myImage"
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

		snapName := "mySnap"
		snapshot, err := img.CreateSnapshot(snapName)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, snapshot.Remove())
		}()

		snapID, err := img.GetSnapID(snapshot.name)
		assert.NoError(t, err)
		snapNameByID, err := img.GetSnapByID(snapID)
		assert.NoError(t, err)
		assert.Equal(t, snapshot.name, snapNameByID)
	})

	t.Run("closedImage", func(t *testing.T) {
		imgName := "myImage"
		closedImg, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, closedImg.Remove())
		}()

		img, err := OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		snapName := "mySnap"
		snapshot, err := img.CreateSnapshot(snapName)
		assert.NoError(t, err)

		snapID, err := img.GetSnapID(snapshot.name)
		assert.NoError(t, err)

		// close the image
		assert.NoError(t, img.Close())

		// try to get the ID again
		_, err = img.GetSnapID(snapshot.name)
		assert.Error(t, err)

		_, err = img.GetSnapByID(snapID)
		assert.Error(t, err)

		// Open the image for snapshot removal
		img, err = OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		snapshot = img.GetSnapshot(snapName)
		assert.NoError(t, snapshot.Remove())
		assert.NoError(t, img.Close())
	})

	t.Run("invalidOptions", func(t *testing.T) {
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

		_, err = img.GetSnapID("")
		assert.Error(t, err)

		_, err = img.GetSnapID("someSnap")
		assert.Error(t, err)

		var snapID uint64
		snapID = 22
		_, err = img.GetSnapByID(snapID)
		assert.Error(t, err)
	})
}
