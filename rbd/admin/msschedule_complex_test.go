//go:build !nautilus
// +build !nautilus

package admin

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/rbd"
)

func skipIfQuincy(t *testing.T) {
	vname := os.Getenv("CEPH_VERSION")
	if vname == "quincy" {
		t.Skipf("disabled on ceph %s", vname)
	}
}

func TestMirrorSnapshotScheduleStatus(t *testing.T) {
	// note: the status function doesn't return anything "useful" unless
	// there's an image in the pool. thus we require an image first.
	ensureDefaultPool(t)
	conn := getConn(t)

	ioctx, err := conn.OpenIOContext(defaultPoolName)
	require.NoError(t, err)
	defer ioctx.Destroy()

	imgName := "img1"
	options := rbd.NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(rbd.ImageOptionOrder, uint64(testImageOrder)))
	err = rbd.CreateImage(ioctx, imgName, testImageSize, options)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, rbd.RemoveImage(ioctx, imgName))
	}()
	img, err := rbd.OpenImage(ioctx, imgName, rbd.NoSnapshot)
	assert.NoError(t, err)
	err = img.MirrorEnable(rbd.ImageMirrorModeSnapshot)
	assert.NoError(t, err)
	assert.NoError(t, img.Close())

	ra := getAdmin(t)
	scheduler := ra.MirrorSnashotSchedule()
	err = scheduler.Add(
		NewLevelSpec(defaultPoolName, "", imgName),
		Interval("1d"),
		NoStartTime)
	assert.NoError(t, err)
	defer func() {
		err = scheduler.Remove(
			NewLevelSpec(defaultPoolName, "", imgName),
			Interval("1d"),
			NoStartTime)
		assert.NoError(t, err)
	}()

	// This is one of those calls that depends on something async inside ceph
	// and doesn't return the "expected result" immediately after the schedule
	// is added. Loop on it for a while checking for the desired condition to
	// become true.
	// Unfortunately, this particular case is quite slow and doesn't
	// seem to be ready until around a minute(!).
	var status []ScheduledImage
	for i := 0; i < 100; i++ {
		status, err = scheduler.Status(
			NewLevelSpec(defaultPoolName, "", imgName))
		assert.NoError(t, err)
		if len(status) == 1 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if assert.Len(t, status, 1) {
		assert.Equal(t, "rbd/img1", status[0].Image)
		// we don't bother asserting the ScheduleTime value because
		// it changes - and it's not worth messing with the system
		// clock just for this kind of test.
	}
}
