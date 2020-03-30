package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRBDError(t *testing.T) {
	err := getError(0)
	assert.NoError(t, err)

	err = getError(-39) // NOTEMPTY (image still has a snapshot)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rbd: ret=39, Directory not empty")

	err = getError(345) // no such errno
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rbd: ret=345")
}
