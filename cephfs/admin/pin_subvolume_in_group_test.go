//go:build !nautilus && ceph_preview

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPinSubVolumeInGroup(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	subvolumegroup := "cephfs_subvol_group"
	err := fsa.CreateSubVolumeGroup(volume, subvolumegroup, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, subvolumegroup)
		assert.NoError(t, err)
	}()

	subvolname := "cephfs_subvol"
	err = fsa.CreateSubVolume(volume, subvolumegroup, subvolname, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, subvolumegroup, subvolname)
		assert.NoError(t, err)
	}()

	var ec errorCode
	_, err = fsa.PinSubVolumeInGroup(volume, subvolumegroup, subvolname, "distributed", "2")
	assert.True(t, errors.As(err, &ec))
	assert.Equal(t, -22, ec.ErrorCode())

	_, err = fsa.PinSubVolumeInGroup(volume, subvolumegroup, subvolname, "distributed", "1")
	assert.NoError(t, err)

	// Also test with NoGroup to verify it behaves like PinSubVolume.
	subvolname2 := "cephfs_subvol_nogroup"
	err = fsa.CreateSubVolume(volume, NoGroup, subvolname2, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, NoGroup, subvolname2)
		assert.NoError(t, err)
	}()

	_, err = fsa.PinSubVolumeInGroup(volume, NoGroup, subvolname2, "distributed", "1")
	assert.NoError(t, err)
}
