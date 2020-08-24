package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoidPtr(t *testing.T) {
	i := uintptr(42)
	j := uintptr(VoidPtr(i))
	assert.Equal(t, i, j)
}
