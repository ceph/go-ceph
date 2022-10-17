package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleCloneStatusFailed = []byte(`{
  "status": {
    "state": "failed",
    "source": {
      "volume": "non-existing-cephfs",
      "subvolume": "subvol1",
      "snapshot": "snap1"
    }
  },
  "failure": {
    "errno": "2",
    "errstr": "No such file or directory"
  }
}`)

// TestParseCloneStatusFailure is heavily based on TestParseCloneStatus, with
// the addition of GetFailure() calls.
func TestParseCloneStatusFailure(t *testing.T) {
	R := newResponse
	t.Run("okPending", func(t *testing.T) {
		status, err := parseCloneStatus(R(sampleCloneStatusPending, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, status) {
			assert.EqualValues(t, ClonePending, status.State)
			assert.EqualValues(t, "cephfs", status.Source.Volume)
			assert.EqualValues(t, "jurrasic", status.Source.SubVolume)
			assert.EqualValues(t, "dinodna", status.Source.Snapshot)
			assert.EqualValues(t, "park", status.Source.Group)
			assert.Nil(t, status.GetFailure())
		}
	})
	t.Run("okInProg", func(t *testing.T) {
		status, err := parseCloneStatus(R(sampleCloneStatusInProg, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, status) {
			assert.EqualValues(t, CloneInProgress, status.State)
			assert.EqualValues(t, "cephfs", status.Source.Volume)
			assert.EqualValues(t, "subvol1", status.Source.SubVolume)
			assert.EqualValues(t, "snap1", status.Source.Snapshot)
			assert.EqualValues(t, "", status.Source.Group)
			assert.Nil(t, status.GetFailure())
		}
	})
	t.Run("failedMissingVolume", func(t *testing.T) {
		status, err := parseCloneStatus(R(sampleCloneStatusFailed, "", nil))
		assert.NoError(t, err)
		if assert.NotNil(t, status) {
			assert.EqualValues(t, CloneFailed, status.State)
			assert.EqualValues(t, "non-existing-cephfs", status.Source.Volume)
			assert.EqualValues(t, "subvol1", status.Source.SubVolume)
			assert.EqualValues(t, "snap1", status.Source.Snapshot)
			assert.EqualValues(t, "", status.Source.Group)
			assert.EqualValues(t, "2", status.GetFailure().Errno)
			assert.EqualValues(t, "No such file or directory", status.GetFailure().ErrStr)
		}
	})
}
