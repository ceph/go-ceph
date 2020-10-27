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
