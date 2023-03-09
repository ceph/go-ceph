//go:build !nautilus && ceph_preview
// +build !nautilus,ceph_preview

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPinSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	subvolname := "cephfs_subvol"
	err := fsa.CreateSubVolume(volume, NoGroup, subvolname, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, NoGroup, subvolname)
		assert.NoError(t, err)
	}()

	var ec errorCode
	_, err = fsa.PinSubVolume(volume, subvolname, "distributed", "2")
	assert.True(t, errors.As(err, &ec))
	assert.Equal(t, -22, ec.ErrorCode())

	_, err = fsa.PinSubVolume(volume, subvolname, "distributed", "1")
	assert.NoError(t, err)
}

func TestPinSubVolumeGroup(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	subvolumegroup := "cephfs_subvol_group"
	err := fsa.CreateSubVolumeGroup(volume, subvolumegroup, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, subvolumegroup)
		assert.NoError(t, err)
	}()

	// mds_export_ephemeral_random_max has a default value of 0.01. EINVAL
	// is returned for an attempt to set a value beyond this config.
	var ec errorCode
	_, err = fsa.PinSubVolumeGroup(volume, subvolumegroup, "random", "0.5")
	assert.True(t, errors.As(err, &ec))
	assert.Equal(t, -22, ec.ErrorCode())

	_, err = fsa.PinSubVolumeGroup(volume, subvolumegroup, "random", "0.01")
	assert.NoError(t, err)
}

type errorCode interface {
	ErrorCode() int
}
