package errutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCephError(t *testing.T) {
	radosErr := cephErrno(107)
	assert.Equal(t, "Transport endpoint is not connected", radosErr.Error())

	cephFSErr := GetError("cephfs", 2)
	assert.Equal(t, "cephfs: ret=2, No such file or directory",
		cephFSErr.Error())
	assert.Equal(t, 2, cephFSErr.(cephError).ErrorCode())

	rbdErr := GetError("rbd", 2)
	assert.True(t, errors.Is(cephFSErr, rbdErr))
	assert.True(t, errors.Unwrap(cephFSErr) == errors.Unwrap(rbdErr))
}
