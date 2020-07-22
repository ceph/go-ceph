package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStatFSRootDir does not assert much about every field as these can vary
// between runs. Similarly, some stats might vary between sub-trees but we
// trust the ceph libs to be correct here and just make sure the wrapper code
// behaves.
func TestStatFSRootDir(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		m := &MountInfo{}
		sfs, err := m.StatFS("/")
		assert.Error(t, err)
		assert.Nil(t, sfs)
	})

	// half the stats as reported by ceph are pretty useless/dummy values.
	// (see src/client/Client.cc)
	// some stuff gets filled in only if a quota is set, but we're not
	// up to that right now, so we don't really check much value-wise.
	t.Run("valid", func(t *testing.T) {
		mount := fsConnect(t)
		defer fsDisconnect(t, mount)

		sfs, err := mount.StatFS("/")
		assert.NoError(t, err)
		assert.NotNil(t, sfs)
		assert.Equal(t, sfs.Namemax, int64(255))
	})
}
