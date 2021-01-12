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
	var data [][]byte
	for _, s := range strs {
		data = append(data, []byte(s))
	}
	iovec := ByteSlicesToIovec(data)
	assert.Equal(t, iovec.data, data)
	assert.Equal(t, iovec.Pointer(), unsafe.Pointer(&iovec.iovec[0]))
	assert.Equal(t, iovec.Len(), len(data))
	for i, iov := range iovec.iovec {
		require.NotNil(t, iov.iov_base)
		assert.Equal(t, int(iov.iov_len), len(data[i]))
		assert.Equal(t, data[i], (*[999]byte)(iov.iov_base)[:int(iov.iov_len):int(iov.iov_len)])
		for j := range data[i] {
			data[i][j] = 0
		}
	}
	for i, b := range data {
		assert.NotEqual(t, string(b), strs[i])
	}
	iovec.SyncToData()
	for i, b := range data {
		assert.Equal(t, string(b), strs[i])
	}
	data[0] = []byte("changed")
	assert.Panics(t, func() { iovec.SyncToData() })
	iovec.Free()
	for _, iov := range iovec.iovec {
		assert.Equal(t, iov.iov_base, unsafe.Pointer(nil))
		assert.Zero(t, iov.iov_len)
	}
	iovec.Free()
}
