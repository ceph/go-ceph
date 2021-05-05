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
	p := iovec.Pointer()
	assert.NotNil(t, p)
	assert.Equal(t, iovec.Len(), len(data))
	assert.Equal(t, p, unsafe.Pointer(&iovec.iovec[0]))
	for i, iov := range iovec.iovec {
		require.NotNil(t, iov.iov_base)
		assert.Equal(t, int(iov.iov_len), len(data[i]))
		assert.Equal(t, data[i], (*[MaxIdx]byte)(iov.iov_base)[:iov.iov_len:iov.iov_len])
	}
	// data didn't change
	for i, b := range data {
		assert.Equal(t, string(b), strs[i])
	}
	// clear iovec buffers
	for _, iov := range iovec.iovec {
		b := (*[MaxIdx]byte)(iov.iov_base)[:iov.iov_len:iov.iov_len]
		for i := range b {
			b[i] = 0
		}
	}
	iovec.Sync()
	// data must be cleared
	for _, b := range data {
		for i := range b {
			assert.Zero(t, b[i])
		}
	}
	iovec.Free()
	for _, iov := range iovec.iovec {
		assert.Equal(t, iov.iov_base, unsafe.Pointer(nil))
		assert.Zero(t, iov.iov_len)
	}
	iovec.Free()
	iovec.Sync()
	iovec.Sync()
	iovec.Free()
}

func BenchmarkIovec(b *testing.B) {
	data := make([][]byte, 64)
	for i := range data {
		data[i] = make([]byte, 1024*64)
	}
	for i := 0; i < b.N; i++ {
		iovec := ByteSlicesToIovec(data)
		iovec.Sync()
		iovec.Free()
	}
}
