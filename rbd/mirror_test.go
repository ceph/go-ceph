// +build !nautilus

// Initially, we're only providing mirroring related functions for octopus as
// that version of ceph deprecated a number of the functions in nautilus. If
// you need mirroring on an earlier supported version of ceph please file an
// issue in our tracker.

package rbd

import (
	"fmt"
	"os"
	"testing"
	"time"

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

func mirrorConfig() string {
	return os.Getenv("MIRROR_CONF")
}

func TestGetGlobalMirrorStatusMirroredPool(t *testing.T) {
	mconfig := mirrorConfig()
	if mconfig == "" {
		t.Skip("no mirror config env var set")
	}
	conn := radosConnect(t)
	// this test assumes the rbd pool already exists and is mirrored
	// this must be set up previously by the CI or manually
	poolName := "rbd"

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	imgName := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, imgName, testImageSize, options)
	require.NoError(t, err)

	defer func() {
		err = RemoveImage(ioctx, imgName)
		assert.NoError(t, err)
	}()

	// this next section is not a t.Run because it must be unconditionally
	// executed. It is wrapped in a func to use defer to close the img.
	func() {
		img, err := OpenImage(ioctx, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)

		mid, err := img.CreateMirrorSnapshot()
		assert.NoError(t, err)
		assert.NotEqual(t, 0, mid)

		// wait for site statuses to get updated
		for i := 0; i < 30; i++ {
			gms, err := img.GetGlobalMirrorStatus()
			assert.NoError(t, err)
			if len(gms.SiteStatuses) > 1 {
				break
			}
			time.Sleep(time.Second)
		}

		gms, err := img.GetGlobalMirrorStatus()
		assert.NoError(t, err)
		assert.NotEqual(t, "", gms.Name)
		assert.NotEqual(t, "", gms.Info.GlobalID)
		assert.Equal(t, gms.Info.State, MirrorImageEnabled)
		assert.Equal(t, gms.Info.Primary, true)
		if assert.Len(t, gms.SiteStatuses, 2) {
			ss1 := gms.SiteStatuses[0]
			assert.Equal(t, "", ss1.MirrorUUID)
			assert.Equal(t, MirrorImageStatusStateStopped, ss1.State, ss1.State)
			assert.Equal(t, "local image is primary", ss1.Description)
			assert.Greater(t, ss1.LastUpdate, int64(0))
			assert.True(t, ss1.Up)
			ls, err := gms.LocalStatus()
			assert.NoError(t, err)
			assert.Equal(t, ss1, ls)

			ss2 := gms.SiteStatuses[1]
			assert.NotEqual(t, "", ss2.MirrorUUID)
			assert.Equal(t, MirrorImageStatusStateReplaying, ss2.State, ss2.State)
			assert.Contains(t, ss2.Description, "replaying,")
			assert.Greater(t, ss2.LastUpdate, int64(0))
			assert.True(t, ss2.Up)
		}
	}()

	// test the results of GetGlobalMirrorStatus using the "other"
	// mirror+pool as a source
	t.Run("fromMirror", func(t *testing.T) {
		conn := radosConnectConfig(t, mconfig)
		ioctx2, err := conn.OpenIOContext(poolName)
		assert.NoError(t, err)
		defer func() {
			ioctx2.Destroy()
		}()

		img, err := OpenImage(ioctx2, imgName, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		// wait for site statuses to get updated
		for i := 0; i < 30; i++ {
			gms, err := img.GetGlobalMirrorStatus()
			assert.NoError(t, err)
			if len(gms.SiteStatuses) > 1 {
				break
			}
			time.Sleep(time.Second)
		}

		gms, err := img.GetGlobalMirrorStatus()
		assert.NoError(t, err)
		assert.NotEqual(t, "", gms.Name)
		assert.NotEqual(t, "", gms.Info.GlobalID)
		assert.Equal(t, gms.Info.State, MirrorImageEnabled)
		assert.Equal(t, gms.Info.Primary, false)
		if assert.Len(t, gms.SiteStatuses, 2) {
			ls, err := gms.LocalStatus()
			assert.NoError(t, err)
			assert.Equal(t, "", ls.MirrorUUID)
			assert.Equal(t, MirrorImageStatusStateReplaying, ls.State, ls.State)
			assert.Contains(t, ls.Description, "replaying,")
			assert.Greater(t, ls.LastUpdate, int64(0))
			assert.True(t, ls.Up)

			assert.Equal(t, ls, gms.SiteStatuses[0])

			ss2 := gms.SiteStatuses[1]
			assert.NotEqual(t, "", ss2.MirrorUUID)
			assert.Equal(t, MirrorImageStatusStateStopped, ss2.State, ss2.State)

			assert.Equal(t, "local image is primary", ss2.Description)
			assert.Greater(t, ss2.LastUpdate, int64(0))
			assert.True(t, ss2.Up)
		}
	})
}

func TestMirrorImageStatusSummary(t *testing.T) {
	t.Run("ioctxNil", func(t *testing.T) {
		assert.Panics(t, func() {
			MirrorImageStatusSummary(nil)
		})
	})
	t.Run("emptyPool", func(t *testing.T) {
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

		ssum, err := MirrorImageStatusSummary(ioctx)
		assert.NoError(t, err)
		assert.Len(t, ssum, 0)
	})
	t.Run("mirroredPool", testMirrorImageStatusSummaryMirroredPool)
}

func testMirrorImageStatusSummaryMirroredPool(t *testing.T) {
	mconfig := mirrorConfig()
	if mconfig == "" {
		t.Skip("no mirror config env var set")
	}
	conn := radosConnect(t)
	// this test assumes the rbd pool already exists and is mirrored
	// this must be set up previously by the CI or manually
	poolName := "rbd"

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	imgBase := GetUUID()
	imgName1 := imgBase + "a"
	imgName2 := imgBase + "b"
	imgName3 := imgBase + "c"
	imgName4 := imgBase + "d"

	options := NewRbdImageOptions()
	assert.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))

	for _, n := range []string{imgName1, imgName2, imgName3, imgName4} {
		err = CreateImage(ioctx, n, testImageSize, options)
		require.NoError(t, err)

		defer func(n string) {
			err = RemoveImage(ioctx, n)
			assert.NoError(t, err)
		}(n)
	}

	mkMirror := func(n string) {
		img, err := OpenImage(ioctx, n, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		err = img.MirrorEnable(ImageMirrorModeSnapshot)
		assert.NoError(t, err)

		mid, err := img.CreateMirrorSnapshot()
		assert.NoError(t, err)
		assert.NotEqual(t, 0, mid)
	}

	checkMirror := func(n string) {
		img, err := OpenImage(ioctx, n, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		// wait for site statuses to get updated
		for i := 0; i < 30; i++ {
			gms, err := img.GetGlobalMirrorStatus()
			assert.NoError(t, err)
			if len(gms.SiteStatuses) > 1 {
				break
			}
			time.Sleep(time.Second)
		}
	}

	for _, n := range []string{imgName1, imgName3} {
		mkMirror(n)
	}
	for _, n := range []string{imgName1, imgName3} {
		checkMirror(n)
	}

	ssum, err := MirrorImageStatusSummary(ioctx)
	assert.NoError(t, err)
	if assert.Len(t, ssum, 1) {
		assert.Contains(t, ssum, MirrorImageStatusStateReplaying)
		assert.GreaterOrEqual(t, ssum[MirrorImageStatusStateReplaying], uint(2))
	}

	// immediately going for status right after enabling mirroring and not
	// waiting for things to settle should give us one unknown status
	mkMirror(imgName2)
	ssum, err = MirrorImageStatusSummary(ioctx)
	assert.NoError(t, err)
	if assert.Len(t, ssum, 2) {
		assert.Contains(t, ssum, MirrorImageStatusStateReplaying)
		assert.GreaterOrEqual(t, ssum[MirrorImageStatusStateReplaying], uint(2))
		assert.Contains(t, ssum, MirrorImageStatusStateUnknown)
		assert.GreaterOrEqual(t, ssum[MirrorImageStatusStateUnknown], uint(1))
	}
}

func TestMirrorSiteName(t *testing.T) {
	t.Run("connNilGet", func(t *testing.T) {
		assert.Panics(t, func() {
			GetMirrorSiteName(nil)
		})
	})
	t.Run("connNilSet", func(t *testing.T) {
		assert.Panics(t, func() {
			SetMirrorSiteName(nil, "foo")
		})
	})
	t.Run("simple", func(t *testing.T) {
		conn := radosConnect(t)
		defer conn.Shutdown()
		err := SetMirrorSiteName(conn, "rbd4eva")
		assert.NoError(t, err)
		n, err := GetMirrorSiteName(conn)
		assert.NoError(t, err)
		assert.Equal(t, "rbd4eva", n)

		err = SetMirrorSiteName(conn, "ceph_a")
		assert.NoError(t, err)
		n, err = GetMirrorSiteName(conn)
		assert.NoError(t, err)
		assert.Equal(t, "ceph_a", n)
	})
	t.Run("twoCluster", func(t *testing.T) {
		mconfig := mirrorConfig()
		if mconfig == "" {
			t.Skip("no mirror config env var set")
		}

		conn1 := radosConnect(t)
		defer conn1.Shutdown()
		conn2 := radosConnectConfig(t, mconfig)
		defer conn2.Shutdown()

		err := SetMirrorSiteName(conn1, "cluster_a")
		assert.NoError(t, err)
		err = SetMirrorSiteName(conn2, "cluster_b")
		assert.NoError(t, err)

		// verify the two conns are properly separate
		n1, err := GetMirrorSiteName(conn1)
		assert.NoError(t, err)
		assert.Equal(t, "cluster_a", n1)

		n2, err := GetMirrorSiteName(conn2)
		assert.NoError(t, err)
		assert.Equal(t, "cluster_b", n2)
	})
}
