//go:build main

package cephfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFallocateModeZeroUnsupported and this test file exists merely to track
// the backports for https://tracker.ceph.com/issues/68026. Once they are
// available with release versions this can probably vanish.
func TestFallocateModeZeroUnsupported(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "file1.txt"
	f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f)
	defer func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	err = f.Fallocate(FallocNoFlag, 0, 10)
	assert.Error(t, err)
	assert.Equal(t, ErrOpNotSupported, err)
}
