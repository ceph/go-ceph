// +build !luminous,!mimic

package admin

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func delay() {
	// ceph seems to do this (partly?) async. So for now, we cheat
	// and sleep a little to make subsequent tests more reliable
	time.Sleep(50 * time.Millisecond)
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

	err = fsa.CreateSubVolume(volume, NoGroup, "deletemev1", nil)
	assert.NoError(t, err)

	lsv, err = fsa.ListSubVolumes(volume, NoGroup)
	assert.NoError(t, err)
	afterCount := len(lsv)
	assert.Equal(t, beforeCount, afterCount-1)

	err = fsa.RemoveSubVolume(volume, NoGroup, "deletemev1")
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
	assert.GreaterOrEqual(t, 2020, vinfo.Ctime.Year())
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
