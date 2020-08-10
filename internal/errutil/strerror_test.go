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

func TestFormatErrorCode(t *testing.T) {
	msg := FormatErrorCode("test", -39)
	assert.Equal(t, msg, "test: ret=-39, Directory not empty")

	msg = FormatErrorCode("test", -5)
	assert.Equal(t, msg, "test: ret=-5, Input/output error")

	msg = FormatErrorCode("boop", 345)
	assert.Equal(t, msg, "boop: ret=345")
}
