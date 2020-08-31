// +build !luminous,!mimic

package admin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func delay() {
	// ceph seems to do this (partly?) async. So for now, we cheat
	// and sleep a little to make subsequent tests more reliable
	time.Sleep(50 * time.Millisecond)
}

func TestCreateSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	type gn struct {
		group string
		name  string
	}
	created := []gn{}
	defer func() {
		for _, c := range created {
			err := fsa.RemoveSubVolume(volume, c.group, c.name)
			assert.NoError(t, err)
			delay()
			if c.group != NoGroup {
				err := fsa.RemoveSubVolumeGroup(volume, c.group)
				assert.NoError(t, err)
			}
		}
	}()

	t.Run("simple", func(t *testing.T) {
		subname := "SubVol1"
		err := fsa.CreateSubVolume(volume, NoGroup, subname, nil)
		assert.NoError(t, err)
		created = append(created, gn{NoGroup, subname})

		lsv, err := fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})

	t.Run("options", func(t *testing.T) {
		subname := "SubVol2"
		o := &SubVolumeOptions{
			Mode: 0777,
			Uid:  200,
			Gid:  200,
		}
		err := fsa.CreateSubVolume(volume, NoGroup, subname, o)
		assert.NoError(t, err)
		created = append(created, gn{NoGroup, subname})

		lsv, err := fsa.ListSubVolumes(volume, NoGroup)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})

	t.Run("withGroup", func(t *testing.T) {
		group := "withGroup1"
		subname := "SubVol3"

		err := fsa.CreateSubVolumeGroup(volume, group, nil)
		assert.NoError(t, err)

		err = fsa.CreateSubVolume(volume, group, subname, nil)
		assert.NoError(t, err)
		created = append(created, gn{group, subname})

		lsv, err := fsa.ListSubVolumes(volume, group)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})

	t.Run("groupAndOptions", func(t *testing.T) {
		group := "withGroup2"
		subname := "SubVol4"
		err := fsa.CreateSubVolumeGroup(volume, group, nil)
		assert.NoError(t, err)

		o := &SubVolumeOptions{
			Size: 5 * gibiByte,
			Mode: 0777,
			Uid:  200,
			Gid:  200,
		}
		err = fsa.CreateSubVolume(volume, group, subname, o)
		assert.NoError(t, err)
		created = append(created, gn{group, subname})

		lsv, err := fsa.ListSubVolumes(volume, group)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(lsv), 1)
		assert.Contains(t, lsv, subname)
	})
}

func TestRemoveSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"

	lsv, err := fsa.ListSubVolumes(volume, NoGroup)
	assert.NoError(t, err)
	beforeCount := len(lsv)

	err = fsa.CreateSubVolume(volume, NoGroup, "deletemev1", nil)
	assert.NoError(t, err)

	lsv, err = fsa.ListSubVolumes(volume, NoGroup)
	assert.NoError(t, err)
	afterCount := len(lsv)
	assert.Equal(t, beforeCount, afterCount-1)

	err = fsa.RemoveSubVolume(volume, NoGroup, "deletemev1")
	assert.NoError(t, err)

	delay()
	lsv, err = fsa.ListSubVolumes(volume, NoGroup)
	assert.NoError(t, err)
	nowCount := len(lsv)
	if !assert.Equal(t, beforeCount, nowCount) {
		// this is a hack for debugging a flapping test
		assert.Equal(t, []string{}, lsv)
	}
}

func TestResizeSubVolume(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "sizedGroup"
	subname := "sizeMe1"

	err := fsa.CreateSubVolumeGroup(volume, group, nil)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeGroup(volume, group)
		assert.NoError(t, err)
	}()

	svopts := &SubVolumeOptions{
		Mode: 0777,
		Size: 20 * gibiByte,
	}
	err = fsa.CreateSubVolume(volume, group, subname, svopts)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, subname)
		assert.NoError(t, err)
	}()

	lsv, err := fsa.ListSubVolumes(volume, group)
	assert.NoError(t, err)
	assert.Contains(t, lsv, subname)

	rr, err := fsa.ResizeSubVolume(volume, group, subname, 30*gibiByte, false)
	assert.NoError(t, err)
	assert.NotNil(t, rr)

	rr, err = fsa.ResizeSubVolume(volume, group, subname, 10*gibiByte, true)
	assert.NoError(t, err)
	assert.NotNil(t, rr)

	rr, err = fsa.ResizeSubVolume(volume, group, subname, Infinite, true)
	assert.NoError(t, err)
	assert.NotNil(t, rr)
}

func TestSubVolumePath(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "svpGroup"
	subname := "svp1"

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

	path, err := fsa.SubVolumePath(volume, group, subname)
	assert.NoError(t, err)
	assert.Contains(t, path, group)
	assert.Contains(t, path, subname)
	assert.NotContains(t, path, "\n")

	// invalid subname
	path, err = fsa.SubVolumePath(volume, group, "oops")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}
