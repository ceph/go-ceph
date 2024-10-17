//go:build main

package admin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCloneProgress(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := "Park"
	subname := "Jurrasic"
	snapname := "dinodna1"
	clonename := "babydino"

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

	err = fsa.CreateSubVolumeSnapshot(volume, group, subname, snapname)
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolumeSnapshot(volume, group, subname, snapname)
		assert.NoError(t, err)
	}()

	err = fsa.CloneSubVolumeSnapshot(
		volume, group, subname, snapname, clonename,
		&CloneOptions{TargetGroup: group})
	assert.NoError(t, err)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, clonename)
		assert.NoError(t, err)
	}()

	wasInProgress := false
	for done := false; !done; {
		status, err := fsa.CloneStatus(volume, group, clonename)
		assert.NoError(t, err)
		assert.NotNil(t, status)
		switch status.State {
		case ClonePending:
		case CloneInProgress:
			wasInProgress = true
			assert.NotNil(t, status.ProgressReport.PercentageCloned)
			assert.NotNil(t, status.ProgressReport.AmountCloned)
			assert.NotNil(t, status.ProgressReport.FilesCloned)
		case CloneComplete:
			done = true
		case CloneFailed:
			t.Fatal("clone failed")
		default:
			t.Fatalf("invalid clone status: %q", status.State)
		}
		time.Sleep(5 * time.Millisecond)
	}
	assert.Equal(t, wasInProgress, true)
}
