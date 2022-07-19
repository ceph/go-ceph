//go:build !(nautilus || octopus) && ceph_preview && ceph_pre_quincy
// +build !nautilus,!octopus,ceph_preview,ceph_pre_quincy

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)

	err = fsa.SetMetadata(volume, group, subname, key, value)
	assert.NoError(t, err)

	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestGetMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)

	err = fsa.SetMetadata(volume, group, subname, key, value)
	assert.NoError(t, err)

	metaValue, err := fsa.GetMetadata(volume, group, subname, key)
	assert.NoError(t, err)
	assert.Equal(t, metaValue, value)

	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestRemoveMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)

	err = fsa.SetMetadata(volume, group, subname, key, value)
	assert.NoError(t, err)

	metaValue, err := fsa.GetMetadata(volume, group, subname, key)
	assert.NoError(t, err)
	assert.Equal(t, metaValue, value)

	err = fsa.RemoveMetadata(volume, group, subname, key)
	assert.NoError(t, err)

	metaValue, err = fsa.GetMetadata(volume, group, subname, key)
	assert.Error(t, err)

	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestForceRemoveMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)

	err = fsa.SetMetadata(volume, group, subname, key, value)
	assert.NoError(t, err)

	metaValue, err := fsa.GetMetadata(volume, group, subname, key)
	assert.NoError(t, err)
	assert.Equal(t, metaValue, value)

	err = fsa.ForceRemoveMetadata(volume, group, subname, key)
	assert.NoError(t, err)

	metaValue, err = fsa.GetMetadata(volume, group, subname, key)
	assert.Error(t, err)

	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestListMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	key1 := "hi1"
	value1 := "hello1"
	key2 := "hi2"
	value2 := "hello2"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)

	err = fsa.SetMetadata(volume, group, subname, key1, value1)
	assert.NoError(t, err)

	err = fsa.SetMetadata(volume, group, subname, key2, value2)
	assert.NoError(t, err)

	metaList, err := fsa.ListMetadata(volume, group, subname)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(metaList), 2)
	assert.Contains(t, metaList, key1)
	assert.Contains(t, metaList, key2)

	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}
