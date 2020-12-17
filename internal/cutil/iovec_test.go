package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIovec(t *testing.T) {
	strs := []string{
		"foo",
		"barbar",
		"bazbazbaz",
	}
	data := make([][]byte, len(strs))
	iovec := ByteSlicesToIovec(data)
	// filling data should also work after construction
	for i, s := range strs {
		data[i] = []byte(s)
	}
	p := iovec.Pointer()
	assert.NotNil(t, p)
	assert.Equal(t, iovec.Len(), len(data))
	assert.Equal(t, p, unsafe.Pointer(&iovec.iovec[0]))
	for i, iov := range iovec.iovec {
		require.NotNil(t, iov.iov_base)
		assert.Equal(t, int(iov.iov_len), len(data[i]))
		assert.Equal(t, unsafe.Pointer(&data[i][0]), iov.iov_base)
	}
	// data didn't change
	for i, b := range data {
		assert.Equal(t, string(b), strs[i])
	}
	data[0] = []byte("changed")
	// changed data is picked up
	p = iovec.Pointer()
	assert.NotNil(t, p)
	assert.Equal(t, iovec.Len(), len(data))
	assert.Equal(t, p, unsafe.Pointer(&iovec.iovec[0]))
	for i, iov := range iovec.iovec {
		require.NotNil(t, iov.iov_base)
		assert.Equal(t, int(iov.iov_len), len(data[i]))
		assert.Equal(t, unsafe.Pointer(&data[i][0]), iov.iov_base)
	}
	iovec.Free()
	for _, iov := range iovec.iovec {
		assert.Equal(t, iov.iov_base, unsafe.Pointer(nil))
		assert.Zero(t, iov.iov_len)
	}
	iovec.Free()
}
