package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleSubVolumeGroupInfo1 = []byte(`
{
    "atime": "2021-04-19 18:02:11",
    "bytes_pcent": "undefined",
    "bytes_quota": "infinite",
    "bytes_used": 0,
    "created_at": "2021-04-19 18:02:11",
    "ctime": "2021-04-19 18:02:11",
    "data_pool": "cephfs_data",
    "gid": 0,
    "mode": 16877,
    "mon_addrs": [
        "127.0.0.1:6789"
    ],
    "mtime": "2021-04-19 18:02:11",
    "uid": 0
}
`)

var sampleSubVolumeGroupInfo2 = []byte(`
{
    "atime": "2021-04-20 10:00:00",
    "bytes_pcent": "0.00",
    "bytes_quota": 10737418240,
    "bytes_used": 1024,
    "created_at": "2021-04-20 10:00:00",
    "ctime": "2021-04-20 10:00:00",
    "data_pool": "cephfs_data",
    "gid": 100,
    "mode": 16877,
    "mon_addrs": [
        "127.0.0.1:6789"
    ],
    "mtime": "2021-04-20 10:00:00",
    "uid": 100
}
`)

var badSampleSubVolumeGroupInfo1 = []byte(`
{
    "bytes_quota": "fishy",
    "uid": 0
}
`)

func TestParseSubVolumeGroupInfo(t *testing.T) {
	R := newResponse
	t.Run("error", func(t *testing.T) {
		_, err := parseSubVolumeGroupInfo(R(nil, "", errors.New("gleep glop")))
		assert.Error(t, err)
		assert.Equal(t, "gleep glop", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseSubVolumeGroupInfo(R(nil, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		info, err := parseSubVolumeGroupInfo(R(sampleSubVolumeGroupInfo1, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t, 0, info.Uid)
			assert.Equal(t, 0, info.Gid)
			assert.Equal(t, "cephfs_data", info.DataPool)
			assert.Equal(t, Infinite, info.BytesQuota)
			assert.Equal(t, ByteCount(0), info.BytesUsed)
			assert.Equal(t, 1, len(info.MonAddrs))
			assert.Equal(t, "127.0.0.1:6789", info.MonAddrs[0])
			assert.Equal(t, 2021, info.CreatedAt.Year())
		}
	})
	t.Run("ok2", func(t *testing.T) {
		info, err := parseSubVolumeGroupInfo(R(sampleSubVolumeGroupInfo2, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, info) {
			assert.Equal(t, 100, info.Uid)
			assert.Equal(t, 100, info.Gid)
			assert.Equal(t, ByteCount(10737418240), info.BytesQuota)
			assert.Equal(t, ByteCount(1024), info.BytesUsed)
		}
	})
	t.Run("invalidBytesQuotaValue", func(t *testing.T) {
		info, err := parseSubVolumeGroupInfo(R(badSampleSubVolumeGroupInfo1, "", nil))
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestSubVolumeGroupInfo(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "infogrp"

	err := fsa.CreateSubVolumeGroup(volume, group, &SubVolumeGroupOptions{
		Uid:  200,
		Gid:  200,
		Mode: 0755,
	})
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	info, err := fsa.SubVolumeGroupInfo(volume, group)
	assert.NoError(t, err)
	if assert.NotNil(t, info) {
		assert.Equal(t, 200, info.Uid)
		assert.Equal(t, 200, info.Gid)
		assert.NotEmpty(t, info.DataPool)
		assert.NotEmpty(t, info.MonAddrs)
	}

	// invalid group name
	info, err = fsa.SubVolumeGroupInfo(volume, "oops")
	assert.Error(t, err)
	assert.Nil(t, info)
}
