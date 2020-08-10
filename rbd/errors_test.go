package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRBDError(t *testing.T) {
	err := getError(0)
	assert.NoError(t, err)

	err = getError(-39) // NOTEMPTY (image still has a snapshot)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rbd: ret=39, Directory not empty")

	errno, ok := err.(interface{ ErrorCode() int })
	assert.True(t, ok)
	require.NotNil(t, errno)
	assert.Equal(t, errno.ErrorCode(), -39)

	err = getError(345) // no such errno
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rbd: ret=345")
}
