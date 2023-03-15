//go:build !(nautilus || octopus) && ceph_preview
// +build !nautilus,!octopus,ceph_preview

package admin

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchVolumeInfo(t *testing.T) {
	fsa := getFSAdmin(t)

	t.Run("SimpleVolume", func(t *testing.T) {
		volume := "cephfs"
		vinfo, err := fsa.FetchVolumeInfo(volume)
		assert.NoError(t, err)
		assert.NotNil(t, vinfo)
		assert.Contains(t, vinfo.MonAddrs[0], "6789")
		assert.Equal(t, "cephfs_data", vinfo.Pools.DataPool[0].Name)
	})

	t.Run("InvalidVolume", func(t *testing.T) {
		volume := "blah"
		var ec ErrCode
		_, err := fsa.FetchVolumeInfo(volume)
		assert.True(t, errors.As(err, &ec))
		assert.Equal(t, -2, ec.ErrorCode())
	})

	t.Run("WithSubvolume", func(t *testing.T) {
		volume := "altfs"
		subvolname := "altfs_subvol"

		err := fsa.CreateSubVolume(volume, NoGroup, subvolname, nil)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolume(volume, NoGroup, subvolname)
			assert.NoError(t, err)
		}()

		vinfo, err := fsa.FetchVolumeInfo(volume)
		assert.NoError(t, err)
		assert.NotNil(t, vinfo)
		assert.EqualValues(t, 0, vinfo.PendingSubvolDels)
		assert.Eventually(t,
			func() bool {
				vinfo, err := fsa.FetchVolumeInfo(volume)
				if !assert.NoError(t, err) {
					return false
				}
				return vinfo.Pools.DataPool[0].Used != 0
			},
			10*time.Second,
			100*time.Millisecond,
			"Data pool size not updated")
	})
}

type ErrCode interface {
	ErrorCode() int
}
