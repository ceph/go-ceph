// +build octopus

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleSubVolumeSnapshoInfo1 = []byte(`
{
    "created_at": "2020-09-11 17:40:12.035792",
    "data_pool": "cephfs_data",
    "has_pending_clones": "no",
    "protected": "yes",
    "size": 0
}
`)

func TestParseSubVolumeSnapshotInfo(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		_, err := parseSubVolumeSnapshotInfo(nil, "", errors.New("flub"))
		assert.Error(t, err)
		assert.Equal(t, "flub", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseSubVolumeSnapshotInfo(nil, "unexpected!", nil)
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseSubVolumeSnapshotInfo([]byte("_XxXxX"), "", nil)
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		info, err := parseSubVolumeSnapshotInfo(sampleSubVolumeSnapshoInfo1, "", nil)
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t, "cephfs_data", info.DataPool)
			assert.EqualValues(t, 0, info.Size)
			assert.Equal(t, 2020, info.CreatedAt.Year())
			assert.Equal(t, "yes", info.Protected)
			assert.Equal(t, "no", info.HasPendingClones)
		}
	})
}

func TestSubVolumeSnapshotInfo(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "20000leagues"
	subname := "poulp"
	snapname1 := "t1"
	snapname2 := "nope"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	svopts := &SubVolumeOptions{
		Mode: 0750,
		Size: 20 * gibiByte,
	}
	err = fsa.CreateSubVolume(volume, group, subname, svopts)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, subname)
		assert.NoError(t, err)
	}()

	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname1)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
	}()

	sinfo, err := fsa.SubVolumeSnapshotInfo(volume, group, subname, snapname1)
	assert.NoError(t, err)
	assert.NotNil(t, sinfo)
	assert.EqualValues(t, 0, sinfo.Size)
	assert.Equal(t, "cephfs_data", sinfo.DataPool)
	assert.GreaterOrEqual(t, 2020, sinfo.CreatedAt.Year())

	sinfo, err = fsa.SubVolumeSnapshotInfo(volume, group, subname, snapname2)
	assert.Error(t, err)
	assert.Nil(t, sinfo)
}
