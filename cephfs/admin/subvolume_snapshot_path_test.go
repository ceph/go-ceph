//go:build !(octopus || pacific || quincy || reef || squid)

package admin

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubVolumeSnapshotPath(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	t.Run("non existing snapshot", func(t *testing.T) {
		subname := "subvol"

		err := fsa.CreateSubVolume(volume, NoGroup, subname, nil)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolume(volume, NoGroup, subname)
			assert.NoError(t, err)
		}()

		_, err = fsa.SubVolumeSnapshotPath(volume, NoGroup, subname, "nosnap")
		assert.Error(t, err)
	})

	t.Run("without group", func(t *testing.T) {
		subname := "subvol1"
		snapname := "snap1"

		err := fsa.CreateSubVolume(volume, NoGroup, subname, nil)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolume(volume, NoGroup, subname)
			assert.NoError(t, err)
		}()

		svpath, err := fsa.SubVolumePath(volume, NoGroup, subname)
		assert.NoError(t, err)
		svuuid := path.Base(svpath)

		err = fsa.CreateSubVolumeSnapshot(volume, NoGroup, subname, snapname)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeSnapshot(volume, NoGroup, subname, snapname)
			assert.NoError(t, err)
		}()

		expSnappath := path.Join("/volumes", "_nogroup", subname, ".snap", snapname, svuuid)
		snappath, err := fsa.SubVolumeSnapshotPath(volume, NoGroup, subname, snapname)
		assert.NoError(t, err)
		assert.Equal(t, expSnappath, snappath)
	})

	t.Run("with group", func(t *testing.T) {
		group := "subvolgroup"
		subname := "subvol2"
		snapname := "snap2"

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

		svpath, err := fsa.SubVolumePath(volume, group, subname)
		assert.NoError(t, err)
		svuuid := path.Base(svpath)

		err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
			assert.NoError(t, err)
		}()

		expSnappath := path.Join("/volumes", group, subname, ".snap", snapname, svuuid)
		snappath, err := fsa.SubVolumeSnapshotPath(volume, group, subname, snapname)
		assert.NoError(t, err)
		assert.Equal(t, expSnappath, snappath)
	})
}
