//go:build ceph_preview

package admin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/rbd"
)

func TestTrashPurgeScheduleStatus(t *testing.T) {
	// note: the status function doesn't return anything "useful" unless
	// there's an image in the trash. thus we require an image first.
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

	img, err := rbd.OpenImage(ioctx, imgName, rbd.NoSnapshot)
	assert.NoError(t, err)

	err = img.Trash(0)
	assert.NoError(t, err)
	assert.NoError(t, img.Close())

	trashList, err := rbd.GetTrashList(ioctx)
	assert.NoError(t, err)
	if !assert.Len(t, trashList, 1) {
		t.Fatal("expected one image in trash")
	}
	defer func() {
		trashList, err = rbd.GetTrashList(ioctx)
		if err == nil && len(trashList) > 0 {
			_ = rbd.TrashRemove(ioctx, trashList[0].Id, true)
		}
	}()

	ra := getAdmin(t)
	scheduler := ra.TrashPurgeSchedule()
	err = scheduler.Add(
		NewLevelSpec(defaultPoolName, "", ""),
		Interval("1m"),
		NoStartTime)
	assert.NoError(t, err)
	defer func() {
		err = scheduler.Remove(
			NewLevelSpec(defaultPoolName, "", ""),
			Interval("1m"),
			NoStartTime)
		assert.NoError(t, err)
	}()

	var status []ScheduledPool
	for i := 0; i < 100; i++ {
		status, err = scheduler.Status(
			NewLevelSpec(defaultPoolName, "", ""))
		assert.NoError(t, err)
		if len(status) > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if assert.Len(t, status, 1) {
		assert.Equal(t, defaultPoolName, status[0].PoolName)
		assert.Equal(t, "", status[0].Namespace)
	}
}
