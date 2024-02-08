//go:build !(nautilus || octopus || pacific || quincy || reef || squid || ceph_pre_squid) && ceph_preview

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSQuiesce(t *testing.T) {
	fsa := getFSAdmin(t)
	volume := "cephfs"
	group := NoGroup
	subvol := "quiesceMe"
	fsa.CreateSubVolume(volume, group, subvol, nil)
	defer func() {
		err := fsa.RemoveSubVolume(volume, group, subvol)
		assert.NoError(t, err)
	}()
	ret, err := fsa.FSQuiesce(volume, group, []string{subvol}, "", nil)
	assert.NoError(t, err)
	require.NoError(t, err)
	for _, val := range ret.Sets {
		assert.Equal(t, 0.0, val.Timeout)
	}
	o := &FSQuiesceOptions{}
	o.Timeout = 10.7
	ret, err = fsa.FSQuiesce(volume, group, []string{subvol}, "", o)
	assert.NoError(t, err)
	for _, val := range ret.Sets {
		assert.Equal(t, 10.7, val.Timeout)
	}

	o.Expiration = 15.2
	ret, err = fsa.FSQuiesce(volume, group, []string{subvol}, "", o)
	assert.NoError(t, err)
	for _, val := range ret.Sets {
		assert.Equal(t, 15.2, val.Expiration)
		assert.Equal(t, 10.7, val.Timeout)
	}

	o.Expiration = 15
	ret, err = fsa.FSQuiesce(volume, group, []string{subvol}, "", o)
	assert.NoError(t, err)
	for _, val := range ret.Sets {
		assert.Equal(t, 15.0, val.Expiration)
		assert.Equal(t, 10.7, val.Timeout)
	}
}
