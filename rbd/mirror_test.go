// +build !nautilus

// Initially, we're only providing mirroring related functions for octopus as
// that version of ceph deprecated a number of the functions in nautilus. If
// you need mirroring on an earlier supported version of ceph please file an
// issue in our tracker.

package rbd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMirrorMode(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	t.Run("mirrorModeDisabled", func(t *testing.T) {
		m, err := GetMirrorMode(ioctx)
		assert.NoError(t, err)
		assert.Equal(t, m, MirrorModeDisabled)
	})
	t.Run("mirrorModeEnabled", func(t *testing.T) {
		err = SetMirrorMode(ioctx, MirrorModeImage)
		require.NoError(t, err)
		m, err := GetMirrorMode(ioctx)
		assert.NoError(t, err)
		assert.Equal(t, m, MirrorModeImage)
	})
	t.Run("ioctxNil", func(t *testing.T) {
		assert.Panics(t, func() {
			GetMirrorMode(nil)
		})
	})

}

func TestMirroring(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	// verify that mirroring is not enabled on this new pool
	m, err := GetMirrorMode(ioctx)
	assert.NoError(t, err)
	assert.Equal(t, m, MirrorModeDisabled)

	// enable per-image mirroring for this pool
	err = SetMirrorMode(ioctx, MirrorModeImage)
	require.NoError(t, err)

	name1 := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name1, testImageSize, options)
	require.NoError(t, err)

	t.Run("enableDisable", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)

		mode, err := img.GetImageMirrorMode()
		assert.NoError(t, err)
		assert.Equal(t, mode, ImageMirrorModeSnapshot)

		err = img.MirrorDisable(false)
		assert.NoError(t, err)
	})
	t.Run("enableDisableInvalid", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.Error(t, err)
		err = img.MirrorDisable(false)
		assert.Error(t, err)
		_, err = img.GetImageMirrorMode()
		assert.Error(t, err)
	})
	t.Run("promoteDemote", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)
		err = img.MirrorDemote()
		assert.NoError(t, err)
		err = img.MirrorPromote(false)
		assert.NoError(t, err)
		err = img.MirrorDisable(false)
		assert.NoError(t, err)
	})
	t.Run("promoteDemoteInvalid", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())

		err = img.MirrorDemote()
		assert.Error(t, err)
		err = img.MirrorPromote(false)
		assert.Error(t, err)
	})
	t.Run("resync", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)
		err = img.MirrorDemote()
		assert.NoError(t, err)
		err = img.MirrorResync()
		assert.NoError(t, err)
		err = img.MirrorDisable(true)
		assert.NoError(t, err)
	})
	t.Run("resyncInvalid", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())

		err = img.MirrorResync()
		assert.Error(t, err)
	})
	t.Run("instanceId", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)
		miid, err := img.MirrorInstanceID()
		// this is not currently testable for the "success" case
		// see also the ceph tree where nothing is asserted except
		// that the error is raised.
		// TODO(?): figure out how to test this
		assert.Error(t, err)
		assert.Equal(t, "", miid)
		err = img.MirrorDisable(false)
		assert.NoError(t, err)
	})
	t.Run("instanceIdInvalid", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())

		_, err = img.MirrorInstanceID()
		assert.Error(t, err)
	})
}

func TestGetMirrorImageInfo(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	// enable per-image mirroring for this pool
	err = SetMirrorMode(ioctx, MirrorModeImage)
	require.NoError(t, err)

	imgName := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, imgName, testImageSize, options)
	require.NoError(t, err)

	t.Run("closedImage", func(t *testing.T) {
		img := GetImage(ioctx, imgName)
		_, err = img.GetMirrorImageInfo()
		assert.Error(t, err)
	})

	t.Run("getInfo", func(t *testing.T) {
		// open image, enable, mirroring.
		img, err := OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)
		mii, err := img.GetMirrorImageInfo()
		assert.NoError(t, err)
		assert.NotNil(t, mii.GlobalID)
		assert.Equal(t, mii.State, MirrorImageEnabled)
		assert.Equal(t, mii.Primary, true)
	})
}

func TestMirrorConstantStrings(t *testing.T) {
	x := []struct {
		s fmt.Stringer
		t string
	}{
		{MirrorModeDisabled, "disabled"},
		{MirrorModeImage, "image"},
		{MirrorModePool, "pool"},
		{MirrorMode(9999), "<unknown>"},
		{ImageMirrorModeJournal, "journal"},
		{ImageMirrorModeSnapshot, "snapshot"},
		{ImageMirrorMode(9999), "<unknown>"},
		{MirrorImageDisabling, "disabling"},
		{MirrorImageEnabled, "enabled"},
		{MirrorImageDisabled, "disabled"},
		{MirrorImageState(9999), "<unknown>"},
	}
	for _, v := range x {
		assert.Equal(t, v.s.String(), v.t)
	}
}

func TestGetGlobalMirrorStatus(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	// enable per-image mirroring for this pool
	err = SetMirrorMode(ioctx, MirrorModeImage)
	require.NoError(t, err)

	imgName := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, imgName, testImageSize, options)
	require.NoError(t, err)

	t.Run("closedImage", func(t *testing.T) {
		img := GetImage(ioctx, imgName)
		_, err = img.GetGlobalMirrorStatus()
		assert.Error(t, err)
	})

	t.Run("getStatus", func(t *testing.T) {
		// open image, enable, mirroring.
		img, err := OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)
		gms, err := img.GetGlobalMirrorStatus()
		assert.NoError(t, err)
		assert.NotEqual(t, "", gms.Name)
		assert.NotEqual(t, "", gms.Info.GlobalID)
		assert.Equal(t, gms.Info.State, MirrorImageEnabled)
		assert.Equal(t, gms.Info.Primary, true)
		if assert.Len(t, gms.SiteStatuses, 1) {
			ss := gms.SiteStatuses[0]
			assert.Equal(t, "", ss.MirrorUUID)
			assert.Equal(t, MirrorImageStatusStateUnknown, ss.State, ss.State)
			assert.Equal(t, "status not found", ss.Description)
			assert.Equal(t, int64(0), ss.LastUpdate)
			assert.False(t, ss.Up)
			ls, err := gms.LocalStatus()
			assert.NoError(t, err)
			assert.Equal(t, ss, ls)
		}
	})
}
