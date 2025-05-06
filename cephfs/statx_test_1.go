package cephfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStatxFieldsRootDir does not assert much about every field
// as these can vary between runs. We exercise the getters but
// can only make "lightweight" assertions here.
func TestStatxFieldsRootDir(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	st, err := mount.Statx("/", StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.NotNil(t, st)

	assert.Equal(t, StatxBasicStats, st.Mask&StatxBasicStats)
	// allow Nlink to be >= 2 in the case that some test(s) don't entirely
	// clean up after themselves or the environment is being used outside
	// of the go-ceph suite only.
	assert.GreaterOrEqual(t, st.Nlink, uint32(2))
	assert.Equal(t, uint32(0), st.Uid)
	assert.Equal(t, uint32(0), st.Gid)
	assert.NotEqual(t, uint16(0), st.Mode)
	assert.Equal(t, uint16(0040000), st.Mode&0040000) // is dir?
	assert.NotEqual(t, Inode(0), st.Inode)
	assert.NotEqual(t, uint64(0), st.Dev)
	assert.Equal(t, uint64(0), st.Rdev)
	assert.Greater(t, st.Ctime.Sec, int64(1588711788))
}
