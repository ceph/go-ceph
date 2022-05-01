package admin

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var shortDuration = 50 * time.Millisecond

func delay() {
	// ceph seems to do this (partly?) async. So for now, we cheat
	// and sleep a little to make subsequent tests more reliable
	time.Sleep(shortDuration)
}

func TestCreateSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	type gn struct {
		group string
		name  string
	}
	created := []gn{}
	defer func() {
		for _, c := range created {
			err := fsa.RemoveSubVolume(volume, c.group, c.name)
			assert.NoError(t, err)
			delay()
			if c.group != NoGroup {
				err := fsa.RemoveSubVolumeGroup(volume, c.group)
				assert.NoError(t, err)
			}
		}
	}()

	t.Run("simple", func(t *testing.T) {
		subname := "SubVol1"
		err := fsa.CreateSubVolume(volume, NoGroup, subname, nil)
		assert.NoError(t, err)
		created = append(created, gn{NoGroup, subname})

		lsv, err := fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})

	t.Run("options", func(t *testing.T) {
		subname := "SubVol2"
		o := &SubVolumeOptions{
			Mode: 0777,
			Uid:  200,
			Gid:  200,
		}
		err := fsa.CreateSubVolume(volume, NoGroup, subname, o)
		assert.NoError(t, err)
		created = append(created, gn{NoGroup, subname})

		lsv, err := fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})

	t.Run("withGroup", func(t *testing.T) {
		group := "withGroup1"
		subname := "SubVol3"

		err := fsa.CreateSubVolumeGroup(volume, group, nil)
		assert.NoError(t, err)

		err = fsa.CreateSubVolume(volume, group, subname, nil)
		assert.NoError(t, err)
		created = append(created, gn{group, subname})

		lsv, err := fsa.ListSubVolumes(volume, group)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})

	t.Run("groupAndOptions", func(t *testing.T) {
		group := "withGroup2"
		subname := "SubVol4"
		err := fsa.CreateSubVolumeGroup(volume, group, nil)
		assert.NoError(t, err)

		o := &SubVolumeOptions{
			Size: 5 * gibiByte,
			Mode: 0777,
			Uid:  200,
			Gid:  200,
		}
		err = fsa.CreateSubVolume(volume, group, subname, o)
		assert.NoError(t, err)
		created = append(created, gn{group, subname})

		lsv, err := fsa.ListSubVolumes(volume, group)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})
}

func TestRemoveSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	lsv, err := fsa.ListSubVolumes(volume, NoGroup)
	assert.NoError(t, err)
	beforeCount := len(lsv)

	removeTest := func(t *testing.T, rm func(string, string, string) error) {
		err = fsa.CreateSubVolume(volume, NoGroup, "deletemev1", nil)
		assert.NoError(t, err)

		lsv, err = fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		afterCount := len(lsv)
		assert.Equal(t, beforeCount, afterCount-1)

		err = rm(volume, NoGroup, "deletemev1")
		assert.NoError(t, err)

		delay()
		lsv, err = fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		nowCount := len(lsv)
		if !assert.Equal(t, beforeCount, nowCount) {
			// this is a hack for debugging a flapping test
			assert.Equal(t, []string{}, lsv)
		}
	}

	t.Run("standard", func(t *testing.T) {
		removeTest(t, fsa.RemoveSubVolume)
	})
	t.Run("force", func(t *testing.T) {
		removeTest(t, fsa.ForceRemoveSubVolume)
	})
	t.Run("withFlagsEmpty", func(t *testing.T) {
		removeTest(t, func(v, g, n string) error {
			return fsa.RemoveSubVolumeWithFlags(v, g, n, SubVolRmFlags{})
		})
	})
	t.Run("withFlagsForce", func(t *testing.T) {
		removeTest(t, func(v, g, n string) error {
			return fsa.RemoveSubVolumeWithFlags(v, g, n, SubVolRmFlags{Force: true})
		})
	})
	t.Run("withFlagsRetainSnaps", func(t *testing.T) {
		removeTest(t, func(v, g, n string) error {
			return fsa.RemoveSubVolumeWithFlags(v, g, n, SubVolRmFlags{RetainSnapshots: true})
		})
	})
	t.Run("retainedSnapshotsTest", func(t *testing.T) {
		subname := "retsnap1"
		snapname := "s1"
		err = fsa.CreateSubVolume(volume, NoGroup, subname, nil)
		assert.NoError(t, err)
		vinfo, err := fsa.SubVolumeInfo(volume, NoGroup, subname)
		assert.NoError(t, err)

		canRetain := false
		for _, f := range vinfo.Features {
			if f == SnapshotRetentionFeature {
				canRetain = true
			}
		}
		if !canRetain {
			err = fsa.RemoveSubVolumeWithFlags(
				volume, NoGroup, subname, SubVolRmFlags{Force: true})
			assert.NoError(t, err)
			t.Skipf("this rest of this test requires snapshot retention on the server side")
		}

		lsv, err = fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		afterCount := len(lsv)
		assert.Equal(t, beforeCount, afterCount-1)
		err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snapname)

		err = fsa.RemoveSubVolumeWithFlags(
			volume, NoGroup, subname, SubVolRmFlags{Force: true})
		assert.Error(t, err)

		err = fsa.RemoveSubVolumeWithFlags(
			volume, NoGroup, subname, SubVolRmFlags{RetainSnapshots: true})
		assert.NoError(t, err)

		delay()
		subInfo, err := fsa.SubVolumeInfo(volume, NoGroup, subname)
		assert.NoError(t, err)
		// If the subvolume is deleted with snapshots retained, then
		// it must have snapshot-retained state.
		assert.Equal(t, subInfo.State, StateSnapRetained)

		err = fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snapname)
		assert.NoError(t, err)

		// The deletion of a subvolume in snapshot-retained state is triggered
		// by the deletion of the last snapshot. It does not need to be
		// explicitly deleted.
		// This may also be why we need to wait longer for the subvolume
		// to be removed from the listing.
		// See also: https://tracker.ceph.com/issues/54625

		assert.Eventually(t,
			func() bool {
				lsv, err := fsa.ListSubVolumes(volume, NoGroup)
				if !assert.NoError(t, err) {
					return false
				}
				return len(lsv) == beforeCount
			},
			2*time.Minute,
			shortDuration,
			"subvolume count did not return to previous value")

	})
}

func TestResizeSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "sizedGroup"
	subname := "sizeMe1"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	svopts := &SubVolumeOptions{
		Mode: 0777,
		Size: 20 * gibiByte,
	}
	err = fsa.CreateSubVolume(volume, group, subname, svopts)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, subname)
		assert.NoError(t, err)
	}()

	lsv, err := fsa.ListSubVolumes(volume, group)
	assert.NoError(t, err)
	assert.Contains(t, lsv, subname)

	rr, err := fsa.ResizeSubVolume(volume, group, subname, 30*gibiByte, false)
	assert.NoError(t, err)
	assert.NotNil(t, rr)

	rr, err = fsa.ResizeSubVolume(volume, group, subname, 10*gibiByte, true)
	assert.NoError(t, err)
	assert.NotNil(t, rr)

	rr, err = fsa.ResizeSubVolume(volume, group, subname, Infinite, true)
	assert.NoError(t, err)
	assert.NotNil(t, rr)
}

func TestSubVolumePath(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "svpGroup"
	subname := "svp1"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, subname)
		assert.NoError(t, err)
	}()

	path, err := fsa.SubVolumePath(volume, group, subname)
	assert.NoError(t, err)
	assert.Contains(t, path, group)
	assert.Contains(t, path, subname)
	assert.NotContains(t, path, "\n")

	// invalid subname
	path, err = fsa.SubVolumePath(volume, group, "oops")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}

var sampleSubVolumeInfo1 = []byte(`
{
    "atime": "2020-08-31 19:53:43",
    "bytes_pcent": "undefined",
    "bytes_quota": "infinite",
    "bytes_used": 0,
    "created_at": "2020-08-31 19:53:43",
    "ctime": "2020-08-31 19:57:15",
    "data_pool": "cephfs_data",
    "gid": 0,
    "mode": 16877,
    "mon_addrs": [
        "127.0.0.1:6789"
    ],
    "mtime": "2020-08-31 19:53:43",
    "path": "/volumes/_nogroup/nibbles/df11be81-a648-4a7b-8549-f28306e3ad93",
    "pool_namespace": "",
    "type": "subvolume",
    "uid": 0
}
`)

var sampleSubVolumeInfo2 = []byte(`
{
    "atime": "2020-09-01 17:49:25",
    "bytes_pcent": "0.00",
    "bytes_quota": 444444,
    "bytes_used": 0,
    "created_at": "2020-09-01 17:49:25",
    "ctime": "2020-09-01 23:49:22",
    "data_pool": "cephfs_data",
    "gid": 0,
    "mode": 16877,
    "mon_addrs": [
        "127.0.0.1:6789"
    ],
    "mtime": "2020-09-01 17:49:25",
    "path": "/volumes/_nogroup/nibbles/d6e062df-7fa0-46ca-872a-9adf728e0e00",
    "pool_namespace": "",
    "type": "subvolume",
    "uid": 0
}
`)

var sampleSubVolumeInfo3 = []byte(`
{
    "atime": "2020-10-02 13:48:17",
    "bytes_pcent": "undefined",
    "bytes_quota": "infinite",
    "bytes_used": 0,
    "created_at": "2020-10-02 13:48:17",
    "ctime": "2020-10-02 13:48:17",
    "data_pool": "cephfs_data",
    "features": [
        "snapshot-clone",
        "snapshot-autoprotect"
    ],
    "gid": 0,
    "mode": 16877,
    "mon_addrs": [
        "127.0.0.1:6789"
    ],
    "mtime": "2020-10-02 13:48:17",
    "path": "/volumes/igotta/boogie/0302e067-e7fb-4d9b-8388-aae46164d8b0",
    "pool_namespace": "",
    "type": "subvolume",
    "uid": 0
}
`)

var sampleSubVolumeInfo4 = []byte(`
{
	"atime": "2020-10-02 13:48:17",
	"bytes_pcent": "undefined",
	"bytes_quota": "infinite",
	"bytes_used": 0,
	"created_at": "2020-10-02 13:48:17",
	"ctime": "2020-10-02 13:48:17",
	"data_pool": "cephfs_data",
	"features": [
		"snapshot-clone",
		"snapshot-autoprotect",
		"snapshot-retention"
	],
	"gid": 0,
	"mode": 16877,
	"mon_addrs": [
		"127.0.0.1:6789"
	],
	"mtime": "2020-10-02 13:48:17",
	"path": "/volumes/igotta/boogie/0302e067-e7fb-4d9b-8388-aae46164d8b0",
	"pool_namespace": "",
	"state": "complete",
	"type": "subvolume",
	"uid": 0
}
`)

var badSampleSubVolumeInfo1 = []byte(`
{
    "bytes_quota": "fishy",
    "uid": 0
}
`)

var badSampleSubVolumeInfo2 = []byte(`
{
    "bytes_quota": true,
    "uid": 0
}
`)

func TestParseSubVolumeInfo(t *testing.T) {
	R := newResponse
	t.Run("error", func(t *testing.T) {
		_, err := parseSubVolumeInfo(R(nil, "", errors.New("gleep glop")))
		assert.Error(t, err)
		assert.Equal(t, "gleep glop", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseSubVolumeInfo(R(nil, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		info, err := parseSubVolumeInfo(R(sampleSubVolumeInfo1, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t,
				"/volumes/_nogroup/nibbles/df11be81-a648-4a7b-8549-f28306e3ad93",
				info.Path)
			assert.Equal(t, 0, info.Uid)
			assert.Equal(t, Infinite, info.BytesQuota)
			assert.Equal(t, 040755, info.Mode)
			assert.Equal(t, 2020, info.Ctime.Year())
			assert.Equal(t, "2020-08-31 19:57:15", info.Ctime.String())
		}
	})
	t.Run("ok2", func(t *testing.T) {
		info, err := parseSubVolumeInfo(R(sampleSubVolumeInfo2, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t,
				"/volumes/_nogroup/nibbles/d6e062df-7fa0-46ca-872a-9adf728e0e00",
				info.Path)
			assert.Equal(t, 0, info.Uid)
			assert.Equal(t, ByteCount(444444), info.BytesQuota)
			assert.Equal(t, 040755, info.Mode)
			assert.Equal(t, 2020, info.Ctime.Year())
			assert.Equal(t, "2020-09-01 23:49:22", info.Ctime.String())
		}
	})
	t.Run("ok3", func(t *testing.T) {
		info, err := parseSubVolumeInfo(R(sampleSubVolumeInfo3, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t,
				"/volumes/igotta/boogie/0302e067-e7fb-4d9b-8388-aae46164d8b0",
				info.Path)
			assert.Equal(t, 0, info.Uid)
			assert.Equal(t, Infinite, info.BytesQuota)
			assert.Equal(t, 040755, info.Mode)
			assert.Equal(t, 2020, info.Ctime.Year())
			assert.Equal(t, "2020-10-02 13:48:17", info.Ctime.String())
			assert.Contains(t, info.Features, SnapshotCloneFeature)
			assert.Contains(t, info.Features, SnapshotAutoprotectFeature)
		}
	})
	t.Run("ok4", func(t *testing.T) {
		info, err := parseSubVolumeInfo(R(sampleSubVolumeInfo4, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t,
				"/volumes/igotta/boogie/0302e067-e7fb-4d9b-8388-aae46164d8b0",
				info.Path)
			assert.Equal(t, 0, info.Uid)
			assert.Contains(t, info.Features, SnapshotRetentionFeature)
			assert.Equal(t, info.State, StateComplete)
		}
	})
	t.Run("invalidBytesQuotaValue", func(t *testing.T) {
		info, err := parseSubVolumeInfo(R(badSampleSubVolumeInfo1, "", nil))
		assert.Error(t, err)
		assert.Nil(t, info)
	})
	t.Run("invalidBytesQuotaType", func(t *testing.T) {
		info, err := parseSubVolumeInfo(R(badSampleSubVolumeInfo2, "", nil))
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestSubVolumeInfo(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "hoagie"
	subname := "grinder"

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

	vinfo, err := fsa.SubVolumeInfo(volume, group, subname)
	assert.NoError(t, err)
	assert.NotNil(t, vinfo)
	assert.Equal(t, 0, vinfo.Uid)
	assert.Equal(t, 20*gibiByte, vinfo.BytesQuota)
	assert.Equal(t, 040750, vinfo.Mode)
	assert.Equal(t, time.Now().Year(), vinfo.Ctime.Year())
	// state field was added with snapshot retention feature.
	canRetain := false
	for _, f := range vinfo.Features {
		if f == SnapshotRetentionFeature {
			canRetain = true
		}
	}
	if canRetain {
		assert.Equal(t, vinfo.State, StateComplete)
	}
}

func TestSubVolumeSnapshots(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "20000leagues"
	subname := "nautilus"
	snapname1 := "ne1"
	snapname2 := "mo2"

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

	t.Run("createAndRemove", func(t *testing.T) {
		err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
		err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
	})

	t.Run("createAndForceRemove", func(t *testing.T) {
		err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
		err := fsa.ForceRemoveSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
	})

	t.Run("listOne", func(t *testing.T) {
		err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname1)
			assert.NoError(t, err)
		}()

		snaps, err := fsa.ListSubVolumeSnapshots(volume, group, subname)
		assert.NoError(t, err)
		assert.Len(t, snaps, 1)
		assert.Contains(t, snaps, snapname1)
	})

	t.Run("listTwo", func(t *testing.T) {
		err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname1)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname1)
			assert.NoError(t, err)
		}()
		err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname2)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname2)
			assert.NoError(t, err)
		}()

		snaps, err := fsa.ListSubVolumeSnapshots(volume, group, subname)
		assert.NoError(t, err)
		assert.Len(t, snaps, 2)
		assert.Contains(t, snaps, snapname1)
		assert.Contains(t, snaps, snapname2)
	})
}

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
	R := newResponse
	t.Run("error", func(t *testing.T) {
		_, err := parseSubVolumeSnapshotInfo(R(nil, "", errors.New("flub")))
		assert.Error(t, err)
		assert.Equal(t, "flub", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseSubVolumeSnapshotInfo(R(nil, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseSubVolumeSnapshotInfo(R([]byte("_XxXxX"), "", nil))
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		info, err := parseSubVolumeSnapshotInfo(R(sampleSubVolumeSnapshoInfo1, "", nil))
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
	assert.Equal(t, time.Now().Year(), sinfo.CreatedAt.Year())

	sinfo, err = fsa.SubVolumeSnapshotInfo(volume, group, subname, snapname2)
	assert.Error(t, err)
	assert.Nil(t, sinfo)
}
