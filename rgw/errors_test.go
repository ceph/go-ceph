package rgw

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRGWError(t *testing.T) {
	err := getError(0)
	assert.NoError(t, err)

	err = getError(-5) // IO error
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rgw: ret=-5, Input/output error")

	errno, ok := err.(interface{ ErrorCode() int })
	assert.True(t, ok)
	require.NotNil(t, errno)
	assert.Equal(t, errno.ErrorCode(), -5)

	err = getError(345) // no such errno
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rgw: ret=345")
}
