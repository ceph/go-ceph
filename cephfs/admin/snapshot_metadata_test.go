//go:build !(nautilus || octopus) && ceph_preview && ceph_pre_quincy
// +build !nautilus,!octopus,ceph_preview,ceph_pre_quincy

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSnapshotMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	snapname := "snap1"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)

	err = fsa.SetSnapshotMetadata(volume, group, subname, snapname, key, value)
	assert.NoError(t, err)

	err = fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestGetSnapshotMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	snapname := "snap1"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)

	err = fsa.SetSnapshotMetadata(volume, group, subname, snapname, key, value)
	assert.NoError(t, err)

	metaValue, err := fsa.GetSnapshotMetadata(volume, group, subname, snapname, key)
	assert.NoError(t, err)
	assert.Equal(t, metaValue, value)

	err = fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestRemoveSnapshotMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	snapname := "snap1"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)

	err = fsa.SetSnapshotMetadata(volume, group, subname, snapname, key, value)
	assert.NoError(t, err)

	metaValue, err := fsa.GetSnapshotMetadata(volume, group, subname, snapname, key)
	assert.NoError(t, err)
	assert.Equal(t, metaValue, value)

	err = fsa.RemoveSnapshotMetadata(volume, group, subname, snapname, key)
	assert.NoError(t, err)

	metaValue, err = fsa.GetSnapshotMetadata(volume, group, subname, snapname, key)
	assert.Error(t, err)

	err = fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestForceRemoveSnapshotMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	snapname := "snap1"
	key := "hi"
	value := "hello"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)

	err = fsa.SetSnapshotMetadata(volume, group, subname, snapname, key, value)
	assert.NoError(t, err)

	metaValue, err := fsa.GetSnapshotMetadata(volume, group, subname, snapname, key)
	assert.NoError(t, err)
	assert.Equal(t, metaValue, value)

	err = fsa.ForceRemoveSnapshotMetadata(volume, group, subname, snapname, key)
	assert.NoError(t, err)

	metaValue, err = fsa.GetSnapshotMetadata(volume, group, subname, snapname, key)
	assert.Error(t, err)

	err = fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}

func TestListSnapshotMetadata(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "group"
	subname := "subVol"
	snapname := "snap1"
	key1 := "hi1"
	value1 := "hello1"
	key2 := "hi2"
	value2 := "hello2"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolume(volume, group, subname, nil)
	assert.NoError(t, err)
	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)

	err = fsa.SetSnapshotMetadata(volume, group, subname, snapname, key1, value1)
	assert.NoError(t, err)
	err = fsa.SetSnapshotMetadata(volume, group, subname, snapname, key2, value2)
	assert.NoError(t, err)

	metaList, err := fsa.ListSnapshotMetadata(volume, group, subname, snapname)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(metaList), 2)
	assert.Contains(t, metaList, key1)
	assert.Contains(t, metaList, key2)

	err = fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolume(volume, group, subname)
	assert.NoError(t, err)
	err = fsa.RemoveSubVolumeGroup(volume, group)
	assert.NoError(t, err)
}
