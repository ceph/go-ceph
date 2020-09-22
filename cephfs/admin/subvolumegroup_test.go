// +build !luminous,!mimic

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSubVolumeGroup(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	created := []string{}
	defer func() {
		for _, name := range created {
			err := fsa.RemoveSubVolumeGroup(volume, name)
			assert.NoError(t, err)
		}
	}()

	t.Run("simple", func(t *testing.T) {
		svgroup := "svg1"
		err := fsa.CreateSubVolumeGroup(volume, svgroup, nil)
		assert.NoError(t, err)
		created = append(created, svgroup)

		lsvg, err := fsa.ListSubVolumeGroups(volume)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsvg), 1)
		assert.Contains(t, lsvg, svgroup)
	})

	t.Run("options1", func(t *testing.T) {
		svgroup := "svg2"
		err := fsa.CreateSubVolumeGroup(volume, svgroup, &SubVolumeGroupOptions{
			Mode: 0777,
		})
		assert.NoError(t, err)
		created = append(created, svgroup)

		lsvg, err := fsa.ListSubVolumeGroups(volume)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsvg), 1)
		assert.Contains(t, lsvg, svgroup)
	})

	t.Run("options2", func(t *testing.T) {
		svgroup := "anotherSVG"
		err := fsa.CreateSubVolumeGroup(volume, svgroup, &SubVolumeGroupOptions{
			Uid:  200,
			Gid:  200,
			Mode: 0771,
			// TODO: test pool_layout... I think its a pool name
		})
		assert.NoError(t, err)
		created = append(created, svgroup)

		lsvg, err := fsa.ListSubVolumeGroups(volume)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsvg), 1)
		assert.Contains(t, lsvg, svgroup)
	})
}

func TestRemoveSubVolumeGroup(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	lsvg, err := fsa.ListSubVolumeGroups(volume)
	assert.NoError(t, err)
	beforeCount := len(lsvg)

	removeTest := func(t *testing.T, rm func(string, string) error) {
		err = fsa.CreateSubVolumeGroup(volume, "deleteme1", nil)
		assert.NoError(t, err)

		lsvg, err = fsa.ListSubVolumeGroups(volume)
		assert.NoError(t, err)
		afterCount := len(lsvg)
		assert.Equal(t, beforeCount, afterCount-1)

		err = rm(volume, "deleteme1")
		assert.NoError(t, err)

		lsvg, err = fsa.ListSubVolumeGroups(volume)
		assert.NoError(t, err)
		nowCount := len(lsvg)
		assert.Equal(t, beforeCount, nowCount)
	}

	t.Run("standard", func(t *testing.T) {
		removeTest(t, fsa.RemoveSubVolumeGroup)
	})
	t.Run("force", func(t *testing.T) {
		removeTest(t, fsa.ForceRemoveSubVolumeGroup)
	})
}

func TestSubVolumeGroupPath(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "grewp"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	path, err := fsa.SubVolumeGroupPath(volume, group)
	assert.NoError(t, err)
	assert.Contains(t, path, "/volumes/"+group)
	assert.NotContains(t, path, "\n")

	// invalid group name
	path, err = fsa.SubVolumeGroupPath(volume, "oops")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}

func TestSubVolumeGroupSnapshots(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "bathyscaphe"
	subname := "trieste"
	snapname1 := "ns1"
	snapname2 := "ns2"

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
		err = fsa.CreateSubVolumeGroupSnapshot(volume, group, snapname1)
		assert.NoError(t, err)
		err := fsa.RemoveSubVolumeGroupSnapshot(volume, group, snapname1)
		assert.NoError(t, err)
	})

	t.Run("createAndForceRemove", func(t *testing.T) {
		err = fsa.CreateSubVolumeGroupSnapshot(volume, group, snapname1)
		assert.NoError(t, err)
		err := fsa.ForceRemoveSubVolumeGroupSnapshot(volume, group, snapname1)
		assert.NoError(t, err)
	})

	t.Run("listOne", func(t *testing.T) {
		err = fsa.CreateSubVolumeGroupSnapshot(volume, group, snapname1)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeGroupSnapshot(volume, group, snapname1)
			assert.NoError(t, err)
		}()

		snaps, err := fsa.ListSubVolumeGroupSnapshots(volume, group)
		assert.NoError(t, err)
		assert.Len(t, snaps, 1)
		assert.Contains(t, snaps, snapname1)
	})

	t.Run("listTwo", func(t *testing.T) {
		err = fsa.CreateSubVolumeGroupSnapshot(volume, group, snapname1)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeGroupSnapshot(volume, group, snapname1)
			assert.NoError(t, err)
		}()
		err = fsa.CreateSubVolumeGroupSnapshot(volume, group, snapname2)
		assert.NoError(t, err)
		defer func() {
			err := fsa.RemoveSubVolumeGroupSnapshot(volume, group, snapname2)
			assert.NoError(t, err)
		}()

		snaps, err := fsa.ListSubVolumeGroupSnapshots(volume, group)
		assert.NoError(t, err)
		assert.Len(t, snaps, 2)
		assert.Contains(t, snaps, snapname1)
		assert.Contains(t, snaps, snapname2)

		// subvolumegroup snaps are reflected in subvolumes (with mangled names)
		snaps, err = fsa.ListSubVolumeSnapshots(volume, group, subname)
		assert.NoError(t, err)
		assert.Len(t, snaps, 2)
	})
}
