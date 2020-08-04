package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListVolumes(t *testing.T) {
	fsa := getFSAdmin(t)

	vl, err := fsa.ListVolumes()
	assert.NoError(t, err)
	assert.Len(t, vl, 1)
	assert.Equal(t, "cephfs", vl[0])
}

func TestVolumeStatus(t *testing.T) {
	fsa := getFSAdmin(t)

	vs, err := fsa.VolumeStatus("cephfs")
	assert.NoError(t, err)
	assert.Contains(t, vs.MDSVersion, "version")
}

var sampleVolumeStatus1 = []byte(`
{
"clients": [{"clients": 1, "fs": "cephfs"}],
"mds_version": "ceph version 15.2.4 (7447c15c6ff58d7fce91843b705a268a1917325c) octopus (stable)",
"mdsmap": [{"dns": 76, "inos": 19, "name": "Z", "rank": 0, "rate": 0.0, "state": "active"}],
"pools": [{"avail": 1017799872, "id": 2, "name": "cephfs_metadata", "type": "metadata", "used": 2204126}, {"avail": 1017799872, "id": 1, "name": "cephfs_data", "type": "data", "used": 0}]
}
`)

func TestParseVolumeStatus(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		_, err := parseVolumeStatus(nil, "", errors.New("bonk"))
		assert.Error(t, err)
		assert.Equal(t, "bonk", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseVolumeStatus(nil, "unexpected!", nil)
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseVolumeStatus([]byte("_XxXxX"), "", nil)
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		s, err := parseVolumeStatus(sampleVolumeStatus1, "", nil)
		assert.NoError(t, err)
		if assert.NotNil(t, s) {
			assert.Contains(t, s.MDSVersion, "ceph version 15.2.4")
			assert.Contains(t, s.MDSVersion, "octopus")
		}
	})
}
