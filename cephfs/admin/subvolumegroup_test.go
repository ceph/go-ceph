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

	err = fsa.CreateSubVolumeGroup(volume, "deleteme1", nil)
	assert.NoError(t, err)

	lsvg, err = fsa.ListSubVolumeGroups(volume)
	assert.NoError(t, err)
	afterCount := len(lsvg)
	assert.Equal(t, beforeCount, afterCount-1)

	err = fsa.RemoveSubVolumeGroup(volume, "deleteme1")
	assert.NoError(t, err)

	lsvg, err = fsa.ListSubVolumeGroups(volume)
	assert.NoError(t, err)
	nowCount := len(lsvg)
	assert.Equal(t, beforeCount, nowCount)
}
