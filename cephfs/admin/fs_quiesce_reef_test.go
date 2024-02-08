//go:build (nautilus || octopus || pacific || quincy || reef || squid || ceph_pre_squid) && ceph_preview

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFSQuiesce(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := NoGroup
	fsa.CreateSubVolume(volume, group, "quiesceMe", nil)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, "quiesceMe")
		assert.NoError(t, err)
	}()
	ret, err := fsa.FSQuiesce(volume, group, []string{"quiesceMe"}, "", nil)
	assert.Nil(t, ret)
	var notImplemented NotImplementedError
	assert.True(t, errors.As(err, &notImplemented))
}
