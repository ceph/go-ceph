package errutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatError(t *testing.T) {
	e, msg := FormatErrno(39)
	assert.Equal(t, 39, e)
	assert.Equal(t, msg, "Directory not empty")

	e, msg = FormatErrno(-5)
	assert.Equal(t, 5, e)
	assert.Equal(t, msg, "Input/output error")

	e, msg = FormatErrno(345)
	assert.Equal(t, 345, e)
	assert.Equal(t, msg, "")
}

func TestStrError(t *testing.T) {
	msg := StrError(39)
	assert.Equal(t, msg, "Directory not empty")

	msg = StrError(-5)
	assert.Equal(t, msg, "Input/output error")

	msg = StrError(345)
	assert.Equal(t, msg, "")
}
