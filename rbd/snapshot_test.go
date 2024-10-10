package rbd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serverVersion string
)

const (
	cephOctopus = "octopus"
	cephPacfic  = "pacific"
	cephQuincy  = "quincy"
	cephReef    = "reef"
	cephSquid   = "squid"
	cephMain    = "main"
)

func init() {
	switch vname := os.Getenv("CEPH_VERSION"); vname {
	case cephOctopus, cephPacfic, cephQuincy, cephReef, cephSquid, cephMain:
		serverVersion = vname
	}
}

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

	err = img.SetSnapshot(snapshot.name)
	assert.Equal(t, err, ErrImageNotOpen)

	err = snapshot.Set()
	assert.Equal(t, err, ErrSnapshotNoName)

	// image can not be removed as the snapshot still exists
	// err = img.Remove()
	// assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestGetSnapTimestamp(t *testing.T) {
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

	t.Run("ClosedImage", func(t *testing.T) {
		imgName := "someImage"
		img, err := Create(ioctx, imgName, testImageSize, testImageOrder, 1)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Remove())
		}()
		var snapID uint64
		snapID = 22
		_, err = img.GetSnapTimestamp(snapID)
		assert.Error(t, err)
		assert.Equal(t, err, ErrImageNotOpen)
	})

	t.Run("invalidSnapID", func(t *testing.T) {
		switch serverVersion {
		case cephOctopus, cephPacfic:
			t.Skip("hits assert due to https://tracker.ceph.com/issues/47287")
		}

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
		_, err = img.GetSnapTimestamp(snapID)
		assert.Error(t, err)
		assert.Equal(t, err, ErrNotFound)
	})

	t.Run("happyPath", func(t *testing.T) {
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

		snapName := "mysnap"
		snapshot, err := img.CreateSnapshot(snapName)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, snapshot.Remove())
		}()

		snapInfo, err := img.GetSnapshotNames()
		assert.NoError(t, err)
		assert.Equal(t, snapName, snapInfo[0].Name)
		snapID := snapInfo[0].Id
		_, err = img.GetSnapTimestamp(snapID)
		assert.NoError(t, err)
	})
}
